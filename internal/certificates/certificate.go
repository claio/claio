package certificates

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"time"
)

type Certificate struct {
	Name           string
	Cert           *x509.Certificate
	CertPrivateKey *rsa.PrivateKey
	CertPEM        string
	KeyPEM         string
}

func CreateCertificate(name string, ca *Certificate) (*Certificate, error) {
	// ca == nil --> create CA else certificate signed by CA
	cert := &x509.Certificate{
		SerialNumber: getSerial(),
		Subject:      getPkixName(),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
	}
	if ca == nil {
		log.Printf("Create CA certificate ...")
		cert.IsCA = true
		cert.BasicConstraintsValid = true
		cert.KeyUsage = x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign
	} else {
		log.Printf("Create certificate for '%s' signed by own CA ...", name)
		cert.IPAddresses = []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback}
		cert.SubjectKeyId = []byte{1, 2, 3, 4, 6}
		cert.KeyUsage = x509.KeyUsageDigitalSignature
	}
	return createCert(name, cert, ca)
}

// --- helpers ------------------------------------------------

func getSerial() *big.Int {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serial, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return big.NewInt(0)
	}
	return serial
}

func getPkixName() pkix.Name {
	return pkix.Name{
		Organization:  []string{"Claio"},
		Country:       []string{"AT"},
		Province:      []string{"STMK"},
		Locality:      []string{"Graz"},
		StreetAddress: []string{"Claio"},
		PostalCode:    []string{"8501"},
	}
}

func createCert(name string, cert *x509.Certificate, ca *Certificate) (*Certificate, error) {
	// create private key
	certPrivateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %s", err)
	}

	// create certificate
	caCert := cert
	caCertPrivateKey := certPrivateKey
	if ca != nil {
		caCert = ca.Cert
		caCertPrivateKey = ca.CertPrivateKey
	}
	certBytes, err := x509.CreateCertificate(rand.Reader, cert, caCert, &certPrivateKey.PublicKey, caCertPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %s", err)
	}

	certPEM := new(bytes.Buffer)
	pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	certPrivKeyPEM := new(bytes.Buffer)
	pem.Encode(certPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certPrivateKey),
	})

	return &Certificate{
		Name:           name,
		Cert:           cert,
		CertPrivateKey: certPrivateKey,
		CertPEM:        certPEM.String(),
		KeyPEM:         certPrivKeyPEM.String(),
	}, nil
}
