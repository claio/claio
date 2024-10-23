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
	"math/big"
	"net"
	"time"
)

type CertificateCreator func(namespace string, name string, ca *Certificate, advertisedName *string, advertisedIp *string) (*Certificate, error)

func NewCaCert(namespace string, name string, ca *Certificate, advertisedName *string, advertisedIp *string) (*Certificate, error) {
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
	return createCert(namespace, name, cert, nil)
}

func NewApiserverCert(namespace string, name string, ca *Certificate, advertisedName *string, advertisedIp *string) (*Certificate, error) {
	ip := net.ParseIP(*advertisedIp)
	if ip == nil {
		return nil, fmt.Errorf("invalid IP address: %s", *advertisedIp)
	}
	cert := &x509.Certificate{
		SerialNumber: getSerial(),
		Subject:      pkix.Name{CommonName: "kube-apiserver"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  []net.IP{net.IPv4(192, 168, 128, 1), net.IPv4(127, 0, 0, 1), ip},
		DNSNames: []string{
			"kubernetes",
			"kubernetes.default",
			"kubernetes.default.svc",
			"kubernetes.default.svc.cluster.local",
			"localhost",
			*advertisedName},
	}

	return createCert(namespace, name, cert, ca)
}

func NewFrontProxyCaCert(namespace string, name string, ca *Certificate, advertisedName *string, advertisedIp *string) (*Certificate, error) {
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
	return createCert(namespace, name, cert, nil)
}

func NewFrontProxyClientCert(namespace string, name string, ca *Certificate, advertisedName *string, advertisedIp *string) (*Certificate, error) {
	cert := &x509.Certificate{
		SerialNumber: getSerial(),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		Subject:      pkix.Name{CommonName: "front-proxy-client"},
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	return createCert(namespace, name, cert, ca)
}

func NewApiserverKubeletClientCert(namespace string, name string, ca *Certificate, advertisedName *string, advertisedIp *string) (*Certificate, error) {
	cert := &x509.Certificate{
		SerialNumber: getSerial(),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		Subject:      pkix.Name{CommonName: "kube-apiserver-kubelet-client", Organization: []string{"system:masters"}},
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	return createCert(namespace, name, cert, ca)
}

func NewKubernetesAdminCert(namespace string, name string, ca *Certificate, advertisedName *string, advertisedIp *string) (*Certificate, error) {
	cert := &x509.Certificate{
		SerialNumber: getSerial(),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		Subject:      pkix.Name{CommonName: "kubernetes-admin", Organization: []string{"system:masters"}},
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	return createCert(namespace, name, cert, ca)
}

func NewKubernetesSchedulerCert(namespace string, name string, ca *Certificate, advertisedName *string, advertisedIp *string) (*Certificate, error) {
	cert := &x509.Certificate{
		SerialNumber: getSerial(),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		Subject:      pkix.Name{CommonName: "system:kube-scheduler"},
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	return createCert(namespace, name, cert, ca)
}

func NewKubernetesControllerCert(namespace string, name string, ca *Certificate, advertisedName *string, advertisedIp *string) (*Certificate, error) {
	cert := &x509.Certificate{
		SerialNumber: getSerial(),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		Subject:      pkix.Name{CommonName: "system:kube-controller-manager"},
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	return createCert(namespace, name, cert, ca)
}

func NewKubernetesKonnectivityCert(namespace string, name string, ca *Certificate, advertisedName *string, advertisedIp *string) (*Certificate, error) {
	cert := &x509.Certificate{
		SerialNumber: getSerial(),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		Subject:      pkix.Name{CommonName: "system:masters", Organization: []string{"system:konnectivity-server"}},
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageCodeSigning},
	}
	return createCert(namespace, name, cert, ca)
}

// we handle the SA RSA pub/priv key pair like a normal cert (to simply coding)
func NewSaRSA(namespace string, name string, ca *Certificate, advertisedName *string, advertisedIp *string) (*Certificate, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %s", err)
	}
	privKeyPEM := new(bytes.Buffer)
	pem.Encode(privKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
	der, err := x509.MarshalPKIXPublicKey(privateKey.Public())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %s", err)
	}
	publicKeyPEM := new(bytes.Buffer)
	pem.Encode(publicKeyPEM, &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: der,
	})
	return &Certificate{
		Name:      name,
		Namespace: namespace,
		Pub:       publicKeyPEM.String(),
		Key:       privKeyPEM.String(),
		Cert:      "",
	}, nil

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
func createCert(namespace string, name string, cert *x509.Certificate, ca *Certificate) (*Certificate, error) {
	// create private key
	certPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %s", err)
	}

	// create certificate
	caCert := cert
	caKey := certPrivateKey
	if ca != nil {
		caCert, err = ca.GetCert()
		if err != nil {
			return nil, fmt.Errorf("failed to get ca certificate: %s", err)
		}
		caKey, err = ca.GetKey()
		if err != nil {
			return nil, fmt.Errorf("failed to get ca private key: %s", err)
		}
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
		Name:      name,
		Namespace: namespace,
		Cert:      certPEM.String(),
		Key:       certPrivKeyPEM.String(),
		Pub:       "",
	}, nil
}
