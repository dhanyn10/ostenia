package ssl

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"ostenia/internal/plugins/utils"
	"path/filepath"
	"runtime"
	"time"
)

var RSAKeySize = 4096

func GenerateRootCA(destDir string) error {
	caPath := filepath.Join(destDir, "ca.crt")
	keyPath := filepath.Join(destDir, "ca.key")

	if _, err := os.Stat(caPath); err == nil {
		return nil // Already exists
	}

	priv, err := rsa.GenerateKey(rand.Reader, RSAKeySize)
	if err != nil {
		return err
	}

	subject := pkix.Name{
		Organization:  []string{"Ostenia Managed CA"},
		Country:       []string{"ID"},
		Province:      []string{"Jakarta"},
		Locality:      []string{"Ostenia"},
		StreetAddress: []string{"Local Development"},
		CommonName:    "Ostenia Root CA",
	}

	// Set expiration to 1 year for Root CA
	expiration := time.Now().AddDate(1, 0, 0)

	pubBytes, _ := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	skid := sha1.Sum(pubBytes) // NOSONAR

	template := x509.Certificate{
		SerialNumber:          big.NewInt(time.Now().Unix()),
		Subject:               subject,
		NotBefore:             time.Now().Add(-1 * time.Hour),
		NotAfter:              expiration,
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		SubjectKeyId:          skid[:],
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return err
	}

	caOut, _ := os.Create(caPath)
	pem.Encode(caOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	caOut.Close()

	keyOut, _ := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	keyOut.Close()

	return TrustRootCA(caPath)
}

func GetRemainingDays(certPath string) (int, error) {
	certData, err := os.ReadFile(certPath)
	if err != nil {
		return 0, err
	}

	block, _ := pem.Decode(certData)
	if block == nil {
		return 0, fmt.Errorf("failed to decode PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return 0, err
	}

	remaining := time.Until(cert.NotAfter)
	days := int(remaining.Hours() / 24)
	if days < 0 {
		return 0, nil
	}
	return days, nil
}

func TrustRootCA(caPath string) error {
	if runtime.GOOS == "windows" {
		// Add to both Local Machine (if admin) and Current User to ensure visibility
		certutilPath := filepath.Join(utils.GetSystemDirectory(), "certutil.exe")
		cmd1 := utils.Executor.Command(certutilPath, "-addstore", "-f", "Root", caPath)
		cmd1.Env = utils.SafeEnv()
		_ = cmd1.Run()
		cmd2 := utils.Executor.Command(certutilPath, "-user", "-addstore", "-f", "Root", caPath)
		cmd2.Env = utils.SafeEnv()
		return cmd2.Run()
	}
	return nil
}

func SignCertificate(caDir string, domain string, destDir string) error {
	certPath := filepath.Join(destDir, domain+".crt")
	if _, err := os.Stat(certPath); err == nil {
		// If cert exists, check if it's new enough (we changed logic, so we might want to force re-sign)
		// For now, let the user delete it via UI (OpenSSL Stop)
		return nil 
	}

	caCertPath := filepath.Join(caDir, "ca.crt")
	caKeyPath := filepath.Join(caDir, "ca.key")

	caCertData, err := os.ReadFile(caCertPath)
	if err != nil { return fmt.Errorf("ca.crt not found: %w", err) }
	caKeyData, err := os.ReadFile(caKeyPath)
	if err != nil { return fmt.Errorf("ca.key not found: %w", err) }

	caBlock, _ := pem.Decode(caCertData)
	if caBlock == nil { return fmt.Errorf("failed to decode ca.crt") }
	caCert, err := x509.ParseCertificate(caBlock.Bytes)
	if err != nil { return fmt.Errorf("failed to parse ca.crt: %w", err) }

	keyBlock, _ := pem.Decode(caKeyData)
	if keyBlock == nil { return fmt.Errorf("failed to decode ca.key") }
	caKey, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil { return fmt.Errorf("failed to parse ca.key: %w", err) }

	size := 2048
	if RSAKeySize < 2048 { size = RSAKeySize }
	priv, err := rsa.GenerateKey(rand.Reader, size)
	if err != nil { return err }

	pubBytes, _ := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	skid := sha1.Sum(pubBytes) // NOSONAR

	template := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			Organization: []string{"Ostenia Local Development"},
			CommonName:   domain,
		},
		DNSNames:              []string{domain, "*." + domain},
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
		NotBefore:             time.Now().Add(-1 * time.Hour),
		NotAfter:              time.Now().AddDate(1, 0, 0), // 1 year
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageDataEncipherment,
		BasicConstraintsValid: true,
		IsCA:                  false,
		SubjectKeyId:          skid[:],
		AuthorityKeyId:        caCert.SubjectKeyId,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, caCert, &priv.PublicKey, caKey)
	if err != nil {
		return err
	}

	certOut, err := os.Create(certPath)
	if err != nil { return err }
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()

	keyOut, err := os.Create(filepath.Join(destDir, domain+".key"))
	if err != nil { return err }
	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	keyOut.Close()

	return nil
}
