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
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

type Certificate struct {
	Key  string
	Pub  string
	Cert string
}

func NewCertificate(namespace, name, key, cert, pub string) *Certificate {
	return &Certificate{
		Key:  key,
		Pub:  pub,
		Cert: cert,
	}
}

func NewCertificateFromSecretData(name string, data map[string][]byte) (*Certificate, error) {
	cert := Certificate{}
	if val, ok := data[name+".key"]; ok {
		cert.Key = string(val)
	} else {
		return nil, fmt.Errorf("missing key for certificate %s", name)
	}
	if val, ok := data[name+".pub"]; ok {
		cert.Pub = string(val)
	} else {
		return nil, fmt.Errorf("missing pub for certificate %s", name)
	}
	if val, ok := data[name+".crt"]; ok {
		cert.Cert = string(val)
	} else {
		return nil, fmt.Errorf("missing crt for certificate %s", name)
	}

	return &cert, nil
}

func (c *Certificate) RawCert() (*x509.Certificate, error) {
	block, _ := pem.Decode([]byte(c.Cert))
	if block == nil {
		return nil, fmt.Errorf("failed to decode certificate from PEM")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %s", err)
	}
	return cert, nil
}

func (c *Certificate) RawKey() (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(c.Key))
	if block == nil {
		return nil, fmt.Errorf("failed to decode key from PEM")
	}
	cert, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %s", err)
	}
	return cert, nil
}

func (c *Certificate) RawPub() (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(c.Pub))
	if block == nil {
		return nil, fmt.Errorf("failed to decode public-key from PEM")
	}
	pub, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public-key: %s", err)
	}
	return pub, nil
}

func (c *Certificate) IsValid() bool {
	return true
}
