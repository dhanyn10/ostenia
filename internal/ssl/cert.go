package ssl

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1" // NOSONAR
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

var (
	GenerateRootCAFunc   = generateRootCA
	GetRemainingDaysFunc = getRemainingDays
	SignCertificateFunc  = signCertificate
)

func GenerateRootCA(destDir string) error {
	return GenerateRootCAFunc(destDir)
}

func GetRemainingDays(certPath string) (int, error) {
	return GetRemainingDaysFunc(certPath)
}

func SignCertificate(caDir string, domain string, destDir string) error {
	return SignCertificateFunc(caDir, domain, destDir)
}

// --- Granular Helper Functions for Mocking and Testing ---

func GeneratePrivateKey(bits int) (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, bits)
}

func EncodeCertToPEM(derBytes []byte) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
}

func EncodeKeyToPEM(key *rsa.PrivateKey) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
}

func WriteToFile(path string, data []byte, perm os.FileMode) error {
	return os.WriteFile(path, data, perm)
}

func LoadCertificate(path string) (*x509.Certificate, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block from %s", path)
	}
	return x509.ParseCertificate(block.Bytes)
}

func LoadPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block from %s", path)
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func GenerateSerialNumber() *big.Int {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		// Fallback to timestamp if random fails, though extremely unlikely
		return big.NewInt(time.Now().Unix())
	}
	return serialNumber
}

func CreateRootCATemplate() x509.Certificate {
	subject := pkix.Name{
		Organization:  []string{"Ostenia Managed CA"},
		Country:       []string{"ID"},
		Province:      []string{"Jakarta"},
		Locality:      []string{"Ostenia"},
		StreetAddress: []string{"Local Development"},
		CommonName:    "Ostenia Root CA",
	}

	return x509.Certificate{
		SerialNumber:          GenerateSerialNumber(),
		Subject:               subject,
		NotBefore:             time.Now().Add(-1 * time.Hour),
		NotAfter:              time.Now().AddDate(1, 0, 0), // 1 year
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
	}
}

func CreateEndEntityTemplate(domain string, caCert *x509.Certificate) x509.Certificate {
	return x509.Certificate{
		SerialNumber: GenerateSerialNumber(),
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
		AuthorityKeyId:        caCert.SubjectKeyId,
	}
}

// --- Main Implementation Functions ---

func generateRootCA(destDir string) error {
	caPath := filepath.Join(destDir, "ca.crt")
	keyPath := filepath.Join(destDir, "ca.key")

	if _, err := os.Stat(caPath); err == nil {
		return nil // Already exists
	}

	priv, err := GeneratePrivateKey(RSAKeySize)
	if err != nil {
		return err
	}

	template := CreateRootCATemplate()

	pubBytes, _ := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	skid := sha1.Sum(pubBytes) // NOSONAR
	template.SubjectKeyId = skid[:]

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return err
	}

	if err := WriteToFile(caPath, EncodeCertToPEM(derBytes), 0644); err != nil {
		return err
	}

	if err := WriteToFile(keyPath, EncodeKeyToPEM(priv), 0600); err != nil {
		return err
	}

	return TrustRootCA(caPath)
}

func getRemainingDays(certPath string) (int, error) {
	cert, err := LoadCertificate(certPath)
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

var TrustRootCAOverride func(caPath string) error

func TrustRootCA(caPath string) error {
	if TrustRootCAOverride != nil {
		return TrustRootCAOverride(caPath)
	}
	if runtime.GOOS == "windows" {
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

func signCertificate(caDir string, domain string, destDir string) error {
	certPath := filepath.Join(destDir, domain+".crt")
	keyPath := filepath.Join(destDir, domain+".key")
	if _, err := os.Stat(certPath); err == nil {
		return nil 
	}

	caCert, err := LoadCertificate(filepath.Join(caDir, "ca.crt"))
	if err != nil { return err }

	caKey, err := LoadPrivateKey(filepath.Join(caDir, "ca.key"))
	if err != nil { return err }

	size := 2048
	if RSAKeySize < 2048 { size = RSAKeySize }
	priv, err := GeneratePrivateKey(size)
	if err != nil { return err }

	template := CreateEndEntityTemplate(domain, caCert)

	pubBytes, _ := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	skid := sha1.Sum(pubBytes) // NOSONAR
	template.SubjectKeyId = skid[:]

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, caCert, &priv.PublicKey, caKey)
	if err != nil {
		return err
	}

	if err := WriteToFile(certPath, EncodeCertToPEM(derBytes), 0644); err != nil {
		return err
	}

	if err := WriteToFile(keyPath, EncodeKeyToPEM(priv), 0600); err != nil {
		return err
	}

	return nil
}
