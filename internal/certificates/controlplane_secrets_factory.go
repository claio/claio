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
	claiov1alpha1 "claio/api/v1alpha1"
	"claio/internal/k8s"
	"claio/internal/utils"
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ControlPlaneSecretsFactory struct {
	k8s k8s.K8s
	res *claiov1alpha1.ControlPlane
	log *utils.Log
}

func NewControlPlaneSecretsFactory(client client.Client, res *claiov1alpha1.ControlPlane, ctx context.Context, scheme *runtime.Scheme, log *utils.Log) *ControlPlaneSecretsFactory {
	return &ControlPlaneSecretsFactory{
		k8s: *k8s.NewK8s(ctx, client, res, scheme),
		res: res,
		log: log,
	}
}

func (s *ControlPlaneSecretsFactory) GetCaCert(namespace string) (*Certificate, error) {
	return s.getCertificate(namespace, "ca", nil, nil, nil, NewCaCert)
}

func (s *ControlPlaneSecretsFactory) GetApiserverCert(namespace string, ca *Certificate) (*Certificate, error) {
	return s.getCertificate(namespace, "apiserver", ca, &s.res.Spec.AdvertiseHost, &s.res.Spec.AdvertiseAddress, NewApiserverCert)
}

func (s *ControlPlaneSecretsFactory) GetApiserverKubeletClientCert(namespace string, ca *Certificate) (*Certificate, error) {
	return s.getCertificate(namespace, "apiserver-kubelet-client", ca, nil, nil, NewApiserverKubeletClientCert)
}

func (s *ControlPlaneSecretsFactory) GetFrontProxyCaCert(namespace string) (*Certificate, error) {
	return s.getCertificate(namespace, "front-proxy-ca", nil, nil, nil, NewFrontProxyCaCert)
}

func (s *ControlPlaneSecretsFactory) GetFrontProxyClientCert(namespace string, ca *Certificate) (*Certificate, error) {
	return s.getCertificate(namespace, "front-proxy-client", ca, nil, nil, NewFrontProxyClientCert)
}

func (s *ControlPlaneSecretsFactory) Get(namespace string, ca *Certificate) (*Certificate, error) {
	return s.getCertificate(namespace, "front-proxy-client", ca, nil, nil, NewFrontProxyClientCert)
}

func (s *ControlPlaneSecretsFactory) GetSaRSA(namespace string) (*Certificate, error) {
	return s.getCertificate(namespace, "sa", nil, nil, nil, NewSaRSA)
}

func (s *ControlPlaneSecretsFactory) CheckSecrets(namespace string) error {
	ca, err := s.GetCaCert(namespace)
	if err != nil {
		return err
	}
	_, err = s.GetApiserverCert(namespace, ca)
	if err != nil {
		return err
	}
	_, err = s.GetApiserverKubeletClientCert(namespace, ca)
	if err != nil {
		return err
	}
	frontProxyCa, err := s.GetFrontProxyCaCert(namespace)
	if err != nil {
		return err
	}
	_, err = s.GetFrontProxyClientCert(namespace, frontProxyCa)
	if err != nil {
		return err
	}
	_, err = s.GetSaRSA(namespace)
	if err != nil {
		return err
	}
	kubeconfig := NewKubeconfigFactoryFromSecretsFactory(s)
	_, err = kubeconfig.GetKubeconfig(namespace)
	if err != nil {
		return err
	}
	return nil
}

// -- helper ----------------------------------------------------------------------

func (s *ControlPlaneSecretsFactory) getCertificate(namespace string, name string, ca *Certificate, advertisedName *string, advertisedIp *string, fn CertificateCreator) (*Certificate, error) {
	cert, err := s.getSecret(namespace, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret %s/%s: %s", namespace, name, err)
	}
	if cert != nil {
		return cert, nil
	}
	s.log.Info("   create certificate and secret: %s", name)
	cert, err = fn(namespace, name, ca, advertisedName, advertisedIp)
	if err != nil {
		return nil, fmt.Errorf("  failed to create certificate %s: %s", name, err)
	}
	if err := s.createCertSecret(namespace, name, cert); err != nil {
		return nil, fmt.Errorf("  failed to create secret %s: %s", name, err)
	}
	return cert, nil
}

func (s *ControlPlaneSecretsFactory) getSecret(namespace string, name string) (*Certificate, error) {
	secretData, err := s.k8s.GetSecret(namespace, name)
	if err != nil {
		return nil, err
	}
	if secretData == nil {
		return nil, nil
	}
	cert := Certificate{
		Name:      name,
		Namespace: namespace,
		Key:       string(secretData[name+".key"]),
		Cert:      "",
		Pub:       "",
	}
	if _, found := secretData[name+".crt"]; found {
		cert.Cert = string(secretData[name+".crt"])
	}
	if _, found := secretData[name+".pub"]; found {
		cert.Pub = string(secretData[name+".pub"])
	}
	return &cert, nil
}

func (s *ControlPlaneSecretsFactory) createCertSecret(namespace string, name string, cert *Certificate) error {
	data := make(map[string][]byte)
	data[name+".key"] = []byte(cert.Key)
	if cert.Cert != "" {
		data[name+".crt"] = []byte(cert.Cert)
	}
	if cert.Pub != "" {
		data[name+".pub"] = []byte(cert.Pub)
	}
	if err := s.k8s.CreateSecret(namespace, name, data); err != nil {
		return err
	}
	return nil
}
