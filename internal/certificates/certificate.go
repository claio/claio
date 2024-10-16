/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
	Name    string
	Cert    *x509.Certificate
	RawCert []byte
	Key     *rsa.PrivateKey
	PEM     *CertificatePEM
}

type CertificatePEM struct {
	Key  string
	Cert string
}

type CertificateCreator func(ca *CertificatePEM, advertisedName *string, advertisedIp *string) (*Certificate, error)

func CreateCaCert(ca *CertificatePEM, advertisedName *string, advertisedIp *string) (*Certificate, error) {
	cert := &x509.Certificate{
		SerialNumber:          big.NewInt(0),
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		Issuer:                pkix.Name{CommonName: "kubernetes"},
		Subject:               pkix.Name{CommonName: "kubernetes"},
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageCertSign,
	}
	return createCert("ca", cert, nil)
}

func CreateKubeApiServerCert(ca *Certificate, dnsnames []string, ipStrings []string) (*Certificate, error) {
	log.Printf("Create 'apiserver' certificate ...")
	dns := append([]string{
		"kubernetes", "kubernetes.default",
		"kubernetes.default.svc", "kubernetes.default.svc.cluster.local"}, dnsnames...)
	ips := []net.IP{net.IPv4(192, 168, 128, 1)}
	for _, ipString := range ipStrings {
		ip := net.ParseIP(ipString)
		if ip == nil {
			return nil, fmt.Errorf("invalid IP address: %s", ipString)
		}
		ips = append(ips, ip)
	}

	cert := &x509.Certificate{
		SerialNumber: getSerial(),
		Subject:      pkix.Name{CommonName: "kube-apiserver"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     dns,
		IPAddresses:  ips,
	}

	return createCert("apiserver", cert, ca)
}

func CreateFrontProxyCaCert() (*Certificate, error) {
	log.Printf("Create 'front-proxy-ca' certificate ...")
	cert := &x509.Certificate{
		SerialNumber:          big.NewInt(0),
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		Issuer:                pkix.Name{CommonName: "front-proxy-ca"},
		Subject:               pkix.Name{CommonName: "front-proxy-ca"},
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageCertSign,
	}
	return createCert("front-proxy-ca", cert, nil)
}

func CreateFrontProxyClientCert(ca *Certificate) (*Certificate, error) {
	log.Printf("Create 'front-proxy-client' certificate ...")
	cert := &x509.Certificate{
		SerialNumber: getSerial(),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		Subject:      pkix.Name{CommonName: "front-proxy-client"},
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	return createCert("front-proxy-client", cert, ca)
}

func CreateApiServerKubeletClientCert(ca *Certificate) (*Certificate, error) {
	log.Printf("Create 'apiserver-kubelet-client' certificate ...")
	cert := &x509.Certificate{
		SerialNumber: getSerial(),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		Subject:      pkix.Name{CommonName: "kube-apiserver-kubelet-client", Organization: []string{"kubeadm:cluster-admins"}},
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	return createCert("apiserver-kubelet-client", cert, ca)
}

// --- helpers -----------------------------------------------------------
func getSerial() *big.Int {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 62)
	serial, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return big.NewInt(1)
	}
	return serial
}

// --- helpers ------------------------------------------------
func createCert(name string, cert *x509.Certificate, ca *Certificate) (*Certificate, error) {
	// create private key
	certPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %s", err)
	}

	// create certificate
	caCert := cert
	caKey := certPrivateKey
	if ca != nil {
		caCert, err = x509.ParseCertificate(ca.RawCert)
		if err != nil {
			return nil, fmt.Errorf("failed to parse CA certificate: %s", err)
		}
		caKey = ca.Key
	}
	certBytes, err := x509.CreateCertificate(rand.Reader, cert, caCert, &certPrivateKey.PublicKey, caKey)
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
		Name:    name,
		Cert:    cert,
		RawCert: certBytes,
		Key:     certPrivateKey,
		PEM:     &CertificatePEM{Cert: certPEM.String(), Key: certPrivKeyPEM.String()},
	}, nil
}
