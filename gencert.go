package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"strings"
	"time"
)

type CertConfig struct {
	HostName  string
	CertFile  string
	KeyFile   string
	OrgName   string
	ValidDays int
}

func main() {
	config := CertConfig{
		HostName:  "localhost",
		CertFile:  "cert.pem",
		KeyFile:   "key.pem",
		OrgName:   "Your Organization",
		ValidDays: 365,
	}

	if err := generateCert(config); err != nil {
		log.Fatalf("Error generating certificate: %v", err)
	}
}

func generateCert(config CertConfig) error {
	// Generate a new private key
	privKey, err := newECDSAKey()
	if err != nil {
		return fmt.Errorf("failed to generate private key: %v", err)
	}

	// Create the certificate template
	serialNum, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return fmt.Errorf("failed to generate serial number: %v", err)
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(time.Duration(config.ValidDays) * 24 * time.Hour)

	template := x509.Certificate{
		SerialNumber: serialNum,
		Subject: pkix.Name{
			Organization: []string{config.OrgName},
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	hosts := strings.Split(config.HostName, ",")
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	// Create the certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privKey.PublicKey, privKey)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %v", err)
	}

	// Save the certificate to file
	certFile, err := os.Create(config.CertFile)
	if err != nil {
		return fmt.Errorf("failed to open %s for writing: %v", config.CertFile, err)
	}
	defer certFile.Close()

	if err := pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certDER}); err != nil {
		return fmt.Errorf("failed to write data to %s: %v", config.CertFile, err)
	}

	// Save the private key to file
	keyFile, err := os.Create(config.KeyFile)
	if err != nil {
		return fmt.Errorf("failed to open %s for writing: %v", config.KeyFile, err)
	}
	defer keyFile.Close()

	privBytes, err := x509.MarshalECPrivateKey(privKey)
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %v", err)
	}

	if err := pem.Encode(keyFile, &pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes}); err != nil {
		return fmt.Errorf("failed to write data to %s: %v", config.KeyFile, err)
	}

	return nil
}

func newECDSAKey() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}