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
	"encoding/pem"
	"fmt"
	"math/big"
)

func getSerial() *big.Int {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 62)
	serial, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return big.NewInt(1)
	}
	return serial
}

func createCert(cert *x509.Certificate, ca *Certificate) (*Certificate, error) {
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

	return &Certificate{
		Key:  privateKeyPEM.String(),
		Pub:  publicKeyPEM.String(),
		Cert: certPEM.String(),
	}, nil
}
