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

package controlplanes

import (
	"bytes"
	"claio/internal/certificates"
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

func getSerial() *big.Int {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 62)
	serial, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return big.NewInt(1)
	}
	return serial
}

func createCert(cert *x509.Certificate, ca *certificates.Certificate) (*certificates.Certificate, error) {
	// create private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %s", err)
	}

	// private PEM
	privateKeyPEM := new(bytes.Buffer)
	pem.Encode(privateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	// public PEM
	der, err := x509.MarshalPKIXPublicKey(privateKey.Public())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %s", err)
	}
	publicKeyPEM := new(bytes.Buffer)
	pem.Encode(publicKeyPEM, &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: der,
	})

	// certificate
	caCert := cert
	caKey := privateKey
	if ca != nil {
		caCert, err = ca.RawCert()
		if err != nil {
			return nil, fmt.Errorf("failed to get ca certificate: %s", err)
		}
		caKey, err = ca.RawKey()
		if err != nil {
			return nil, fmt.Errorf("failed to get ca private key: %s", err)
		}
	}
	certBytes, err := x509.CreateCertificate(rand.Reader, cert, caCert, &privateKey.PublicKey, caKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %s", err)
	}
	certPEM := new(bytes.Buffer)
	pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	return &certificates.Certificate{
		Key:  privateKeyPEM.String(),
		Pub:  publicKeyPEM.String(),
		Cert: certPEM.String(),
	}, nil
}

func newCaCert(ca *certificates.Certificate, advertisedName *string, advertisedIp *string) (*certificates.Certificate, error) {
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
	return createCert(cert, nil)
}

func newApiserverCert(ca *certificates.Certificate, advertisedName *string, advertisedIp *string) (*certificates.Certificate, error) {
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

	return createCert(cert, ca)
}

func newFrontProxyCaCert(ca *certificates.Certificate, advertisedName *string, advertisedIp *string) (*certificates.Certificate, error) {
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
	return createCert(cert, nil)
}

func newFrontProxyClientCert(ca *certificates.Certificate, advertisedName *string, advertisedIp *string) (*certificates.Certificate, error) {
	cert := &x509.Certificate{
		SerialNumber: getSerial(),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		Subject:      pkix.Name{CommonName: "front-proxy-client"},
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	return createCert(cert, ca)
}

func newApiserverKubeletClientCert(ca *certificates.Certificate, advertisedName *string, advertisedIp *string) (*certificates.Certificate, error) {
	cert := &x509.Certificate{
		SerialNumber: getSerial(),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		Subject:      pkix.Name{CommonName: "kube-apiserver-kubelet-client", Organization: []string{"system:masters"}},
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	return createCert(cert, ca)
}

func newKubernetesAdminCert(ca *certificates.Certificate, _, _ *string) (*certificates.Certificate, error) {
	cert := &x509.Certificate{
		SerialNumber: getSerial(),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		Subject:      pkix.Name{CommonName: "kubernetes-admin", Organization: []string{"system:masters"}},
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	return createCert(cert, ca)
}

func newSaCert(ca *certificates.Certificate, advertisedName *string, advertisedIp *string) (*certificates.Certificate, error) {
	cert := &x509.Certificate{
		SerialNumber: getSerial(),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		Subject:      pkix.Name{CommonName: "SA"},
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
	}
	return createCert(cert, ca)
}

func (c *ControlPlane) getCertificateSecret(name string) (*certificates.Certificate, error) {
	secretData, err := c.GetSecret(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret %s/%s: %v", c.Namespace(), name, err)
	}
	if secretData == nil {
		return nil, nil
	}
	cert, err := certificates.NewCertificateFromSecretData(name, secretData)
	if err != nil {
		return nil, err
	}
	return cert, nil
}

func (c *ControlPlane) createCertificateSecret(name string, cert *certificates.Certificate) error {
	data := make(map[string][]byte)
	data[name+".key"] = []byte(cert.Key)
	data[name+".crt"] = []byte(cert.Cert)
	data[name+".pub"] = []byte(cert.Pub)
	if err := c.CreateSecret(name, data); err != nil {
		return err
	}
	return nil
}

func (c *ControlPlane) getCertificate(name, caName string, fn CertificateCreator, forceCreate bool) (*certificates.Certificate, bool, error) {
	cert, err := c.getCertificateSecret(name)
	if err != nil {
		return nil, false, err
	}
	if cert != nil {
		if !forceCreate && cert.IsValid() {
			return cert, false, nil
		}
		c.LogInfo("delete old/invalid secret: %s", name)
		if err := c.DeleteSecret(name); err != nil {
			return nil, true, fmt.Errorf("failed to delete invalid secret %s: %s", name, err)
		}
	}
	c.LogInfo("create certificate: %s", name)
	if caName == "" {
		// a CA
		cert, err = fn(nil, nil, nil)
	} else {
		host := c.Object.Spec.AdvertiseHost
		ip := c.Object.Spec.AdvertiseAddress
		ca, err1 := c.getCertificateSecret(caName)
		if err1 != nil {
			return nil, true, fmt.Errorf("failed to get CA (as secret) %s: %s", caName, err1)
		}
		cert, err = fn(ca, &host, &ip)
	}
	if err != nil {
		return nil, true, fmt.Errorf("failed to create certificate %s: %s", name, err)
	}
	if err := c.createCertificateSecret(name, cert); err != nil {
		return nil, true, fmt.Errorf("failed to create secret %s: %s", name, err)
	}
	return cert, true, nil
}

type CertificateCreator func(ca *certificates.Certificate, advertisedName, advertisedIp *string) (*certificates.Certificate, error)

// ----------------------------------------------------------------

func (c *ControlPlane) GetCaCert(forceCreate bool) (*certificates.Certificate, bool, error) {
	return c.getCertificate("ca", "", newCaCert, forceCreate)
}

func (c *ControlPlane) GetApiserverCert(forceCreate bool) (*certificates.Certificate, bool, error) {
	return c.getCertificate("apiserver", "ca", newApiserverCert, forceCreate)
}

func (c *ControlPlane) GetApiserverKubeletClientCert(forceCreate bool) (*certificates.Certificate, bool, error) {
	return c.getCertificate("apiserver-kubelet-client", "ca", newApiserverKubeletClientCert, forceCreate)
}

func (c *ControlPlane) GetFrontProxyCaCert(forceCreate bool) (*certificates.Certificate, bool, error) {
	return c.getCertificate("front-proxy-ca", "", newFrontProxyCaCert, forceCreate)
}

func (c *ControlPlane) GetFrontProxyClientCert(forceCreate bool) (*certificates.Certificate, bool, error) {
	return c.getCertificate("front-proxy-client", "front-proxy-ca", newFrontProxyClientCert, forceCreate)
}

func (c *ControlPlane) GetSaCert(forceCreate bool) (*certificates.Certificate, bool, error) {
	return c.getCertificate("sa", "", newSaCert, forceCreate)
}

func (c *ControlPlane) reconcileCertificates() (bool, bool, error) {
	c.LogHeader("check secrets ...")
	// ca
	_, caChanged, err := c.GetCaCert(false)
	if err != nil {
		return false, false, fmt.Errorf("failed to get ca")
	}
	// apiserver (force renew if CA has changed)
	_, certChanged, err := c.GetApiserverCert(caChanged)
	if err != nil {
		return caChanged, false, fmt.Errorf("failed to get apiserver")
	}
	// apiserver-kubelet-client (force renew if CA has changed)
	_, changed, err := c.GetApiserverKubeletClientCert(caChanged)
	if err != nil {
		return caChanged, false, fmt.Errorf("failed to get apiserver-kubelet-client")
	}
	certChanged = changed || certChanged
	// front-proxy-ca
	_, frontProxyCaChanged, err := c.GetFrontProxyCaCert(false)
	if err != nil {
		return caChanged, false, fmt.Errorf("failed to get front-proxy-ca")
	}
	certChanged = frontProxyCaChanged || certChanged
	// front-proxy-client
	_, changed, err = c.GetFrontProxyClientCert(frontProxyCaChanged)
	if err != nil {
		return caChanged, false, fmt.Errorf("failed to get front-proxy-client")
	}
	certChanged = changed || certChanged
	// sa
	_, changed, err = c.GetSaCert(false)
	if err != nil {
		return caChanged, false, fmt.Errorf("failed to get sa")
	}
	certChanged = changed || certChanged

	return caChanged, certChanged, nil
}
