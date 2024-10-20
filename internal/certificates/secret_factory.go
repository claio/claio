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
	"claio/internal/utils"
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SecretFactory struct {
	client client.Client
	res    *claiov1alpha1.ControlPlane
	ctx    context.Context
	scheme *runtime.Scheme
	log    *utils.Log
}

func NewSecretFactory(client *client.Client, res *claiov1alpha1.ControlPlane, ctx *context.Context, scheme *runtime.Scheme, log *utils.Log) *SecretFactory {
	return &SecretFactory{
		client: *client,
		res:    res,
		ctx:    *ctx,
		scheme: scheme,
		log:    log,
	}
}

func (s *SecretFactory) GetCaCert(namespace string) (*Certificate, error) {
	return s.getCertificate(namespace, "ca", nil, nil, nil, NewCaCert)
}

func (s *SecretFactory) GetApiserverCert(namespace string, ca *Certificate) (*Certificate, error) {
	return s.getCertificate(namespace, "apiserver", ca, &s.res.Spec.AdvertiseHost, &s.res.Spec.AdvertiseAddress, NewApiserverCert)
}

func (s *SecretFactory) GetApiserverKubeletClientCert(namespace string, ca *Certificate) (*Certificate, error) {
	return s.getCertificate(namespace, "apiserver-kubelet-client", ca, nil, nil, NewApiserverKubeletClientCert)
}

func (s *SecretFactory) GetFrontProxyCaCert(namespace string) (*Certificate, error) {
	return s.getCertificate(namespace, "front-proxy-ca", nil, nil, nil, NewFrontProxyCaCert)
}

func (s *SecretFactory) GetFrontProxyClientCert(namespace string, ca *Certificate) (*Certificate, error) {
	return s.getCertificate(namespace, "front-proxy-client", ca, nil, nil, NewFrontProxyClientCert)
}

func (s *SecretFactory) GetSaRSA(namespace string) (*Certificate, error) {
	return s.getCertificate(namespace, "sa", nil, nil, nil, NewSaRSA)
}

func (s *SecretFactory) CheckSecrets(namespace string) error {
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
	return nil
}

// -- helper ----------------------------------------------------------------------

func (s *SecretFactory) getCertificate(namespace string, name string, ca *Certificate, advertisedName *string, advertisedIp *string, fn CertificateCreator) (*Certificate, error) {
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
	if err := s.createSecret(namespace, name, cert); err != nil {
		return nil, fmt.Errorf("  failed to create secret %s: %s", name, err)
	}
	return cert, nil
}

func (s *SecretFactory) getSecret(namespace string, name string) (*Certificate, error) {
	secret := &corev1.Secret{}
	if err := s.client.Get(
		s.ctx,
		client.ObjectKey{
			Namespace: namespace,
			Name:      name,
		},
		secret,
	); err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	cert := Certificate{
		Name:      name,
		Namespace: namespace,
		Key:       string(secret.Data[name+".key"]),
		Cert:      "",
		Pub:       "",
	}
	if _, found := secret.Data[name+".crt"]; found {
		cert.Cert = string(secret.Data[name+".crt"])
	}
	if _, found := secret.Data[name+".pub"]; found {
		cert.Pub = string(secret.Data[name+".pub"])
	}
	return &cert, nil
}

func (s *SecretFactory) createSecret(namespace string, name string, cert *Certificate) error {
	data := make(map[string][]byte)
	data[name+".key"] = []byte(cert.Key)
	if cert.Cert != "" {
		data[name+".crt"] = []byte(cert.Cert)
	}
	if cert.Pub != "" {
		data[name+".pub"] = []byte(cert.Pub)
	}
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: data,
	}
	if err := ctrl.SetControllerReference(s.res, secret, s.scheme); err != nil {
		return fmt.Errorf("   cannot set owner-reference on secret %s/%s: %s", namespace, name, err)
	}
	if err := s.client.Create(s.ctx, secret); err != nil {
		return fmt.Errorf("  failed to create secret %s/%s: %s", namespace, name, err)
	}
	return nil
}
