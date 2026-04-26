package ssl

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

func GenerateRootCA(destDir string) error {
	caPath := filepath.Join(destDir, "ca.crt")
	keyPath := filepath.Join(destDir, "ca.key")

	if _, err := os.Stat(caPath); err == nil {
		return nil // Already exists
	}

	priv, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}

	subject := pkix.Name{
		Organization: []string{"Ostenia Managed CA"},
		CommonName:   "Ostenia Root CA",
	}

	// Set expiration to 30 days
	expiration := time.Now().AddDate(0, 0, 30)

	template := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               subject,
		NotBefore:             time.Now(),
		NotAfter:              expiration,
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
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
		// certutil -addstore -f "Root" ca.crt
		cmd := exec.Command("certutil", "-addstore", "-f", "Root", caPath)
		return cmd.Run()
	}
	return nil
}

func SignCertificate(caDir string, domain string, destDir string) error {
	caCertPath := filepath.Join(caDir, "ca.crt")
	caKeyPath := filepath.Join(caDir, "ca.key")

	caCertData, _ := os.ReadFile(caCertPath)
	caKeyData, _ := os.ReadFile(caKeyPath)

	block, _ := pem.Decode(caCertData)
	caCert, _ := x509.ParseCertificate(block.Bytes)

	block, _ = pem.Decode(caKeyData)
	caKey, _ := x509.ParsePKCS1PrivateKey(block.Bytes)

	priv, _ := rsa.GenerateKey(rand.Reader, 2048)

	template := x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			CommonName: domain,
		},
		DNSNames:    []string{domain, "*." + domain},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(0, 0, 30), // Consistent 30 days
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, caCert, &priv.PublicKey, caKey)
	if err != nil {
		return err
	}

	certOut, _ := os.Create(filepath.Join(destDir, domain+".crt"))
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()

	keyOut, _ := os.Create(filepath.Join(destDir, domain+".key"))
	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	keyOut.Close()

	return nil
}
