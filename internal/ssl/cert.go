package ssl

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
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

	template := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               subject,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
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

func TrustRootCA(caPath string) error {
	if runtime.GOOS == "windows" {
		// certutil -addstore -f "Root" ca.crt
		cmd := exec.Command("certutil", "-addstore", "-f", "Root", caPath)
		return cmd.Run()
	}
	// For other OS, user might need to do it manually or we implement it later
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
		NotAfter:    time.Now().AddDate(1, 0, 0),
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
