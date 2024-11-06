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
	"claio/internal/resources/controlplanes"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"net"
	"time"
)

type Factory struct {
	Factory *controlplanes.Factory
}

func NewFactory(factory *controlplanes.Factory) *Factory {
	return &Factory{
		Factory: factory,
	}
}

func NewCaCert(ca *Certificate, advertisedName *string, advertisedIp *string) (*Certificate, error) {
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

func NewApiserverCert(ca *Certificate, advertisedName *string, advertisedIp *string) (*Certificate, error) {
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

func NewFrontProxyCaCert(ca *Certificate, advertisedName *string, advertisedIp *string) (*Certificate, error) {
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

func NewFrontProxyClientCert(ca *Certificate, advertisedName *string, advertisedIp *string) (*Certificate, error) {
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

func NewApiserverKubeletClientCert(ca *Certificate, advertisedName *string, advertisedIp *string) (*Certificate, error) {
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

func NewKubernetesAdminCert(ca *Certificate, advertisedName *string, advertisedIp *string) (*Certificate, error) {
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

func NewKubernetesSchedulerCert(ca *Certificate, advertisedName *string, advertisedIp *string) (*Certificate, error) {
	cert := &x509.Certificate{
		SerialNumber: getSerial(),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		Subject:      pkix.Name{CommonName: "system:kube-scheduler"},
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	return createCert(cert, ca)
}

func NewKubernetesControllerCert(ca *Certificate, advertisedName *string, advertisedIp *string) (*Certificate, error) {
	cert := &x509.Certificate{
		SerialNumber: getSerial(),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		Subject:      pkix.Name{CommonName: "system:kube-controller-manager"},
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	return createCert(cert, ca)
}

func NewKubernetesKonnectivityCert(ca *Certificate, advertisedName *string, advertisedIp *string) (*Certificate, error) {
	cert := &x509.Certificate{
		SerialNumber: getSerial(),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		Subject:      pkix.Name{CommonName: "system:masters", Organization: []string{"system:konnectivity-server"}},
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageCodeSigning},
	}
	return createCert(cert, ca)
}

func NewSaCert(ca *Certificate, advertisedName *string, advertisedIp *string) (*Certificate, error) {
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
