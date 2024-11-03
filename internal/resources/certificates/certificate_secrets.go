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
	"fmt"
)

// --- private helpers -------------------------------------------------

func (s *CertificateFactory) GetCertificateSecret(name string) (*Certificate, error) {
	secretData, err := s.Factory.KubernetesClient.GetSecret(s.Factory.Namespace, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret %s/%s: %v", s.Factory.Namespace, name, err)
	}
	if secretData == nil {
		return nil, nil
	}
	cert, err := NewCertificateFromSecretData(name, secretData)
	if err != nil {
		return nil, err
	}
	return cert, nil
}

func (s *CertificateFactory) createCertificateSecret(name string, cert *Certificate) error {
	data := make(map[string][]byte)
	data[name+".key"] = []byte(cert.Key)
	data[name+".crt"] = []byte(cert.Cert)
	data[name+".pub"] = []byte(cert.Pub)
	if err := s.Factory.KubernetesClient.CreateSecret(s.Factory.Namespace, name, data); err != nil {
		return err
	}
	return nil
}

func (s *CertificateFactory) getCertificate(name, caName string, fn CertificateCreator, forceCreate bool) (*Certificate, bool, error) {
	log := s.Factory.Base.Logger(2)
	cert, err := s.GetCertificateSecret(name)
	if err != nil {
		return nil, false, err
	}
	if cert != nil {
		if !forceCreate && cert.IsValid() {
			return cert, false, nil
		}
		log.Info("   delete old/invalid secret: %s", name)
		if err := s.Factory.KubernetesClient.DeleteSecret(s.Factory.Namespace, name); err != nil {
			return nil, true, fmt.Errorf("  failed to delete invalid secret %s: %s", name, err)
		}
	}
	log.Info("   create certificate and secret: %s", name)
	if caName == "" {
		// a CA
		cert, err = fn(nil, nil, nil)
	} else {
		host := s.Factory.Spec.AdvertiseHost
		ip := s.Factory.Spec.AdvertiseAddress
		ca, err1 := s.GetCertificateSecret(caName)
		if err1 != nil {
			return nil, true, fmt.Errorf("  failed to get CA (as secret) %s: %s", caName, err1)
		}
		cert, err = fn(ca, &host, &ip)
	}
	if err != nil {
		return nil, true, fmt.Errorf("  failed to create certificate %s: %s", name, err)
	}
	if err := s.createCertificateSecret(name, cert); err != nil {
		return nil, true, fmt.Errorf("  failed to create secret %s: %s", name, err)
	}
	return cert, true, nil
}

type CertificateCreator func(ca *Certificate, advertisedName, advertisedIp *string) (*Certificate, error)

// ----------------------------------------------------------------

func (s *CertificateFactory) GetCa(forceCreate bool) (*Certificate, bool, error) {
	return s.getCertificate("ca", "", NewCaCert, forceCreate)
}

func (s *CertificateFactory) GetApiserver(forceCreate bool) (*Certificate, bool, error) {
	return s.getCertificate("apiserver", "ca", NewApiserverCert, forceCreate)
}

func (s *CertificateFactory) GetApiserverKubeletClient(forceCreate bool) (*Certificate, bool, error) {
	return s.getCertificate("apiserver-kubelet-client", "ca", NewApiserverKubeletClientCert, forceCreate)
}

func (s *CertificateFactory) GetFrontProxyCa(forceCreate bool) (*Certificate, bool, error) {
	return s.getCertificate("front-proxy-ca", "", NewFrontProxyCaCert, forceCreate)
}

func (s *CertificateFactory) GetFrontProxyClient(forceCreate bool) (*Certificate, bool, error) {
	return s.getCertificate("front-proxy-client", "front-proxy-ca", NewFrontProxyClientCert, forceCreate)
}

func (s *CertificateFactory) GetSa(forceCreate bool) (*Certificate, bool, error) {
	return s.getCertificate("sa", "", NewSaCert, forceCreate)
}
