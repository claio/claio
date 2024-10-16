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
	"context"

	claiov1alpha1 "claio/api/v1alpha1"
	"crypto/x509"
	"fmt"

	"github.com/k0kubun/pp"
	"github.com/smallstep/certinfo"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getSecret(namespace string, name string, c client.Client, ctx context.Context) (*corev1.Secret, error) {
	secret := &corev1.Secret{}
	if err := c.Get(
		ctx,
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
	return secret, nil
}

func CheckSecrets(namespace string, spec claiov1alpha1.ControlPlaneSpec, c client.Client, ctx context.Context) error {
	caSecret, err := getSecret(namespace, "ca", c, ctx)
	if err != nil {
		return err
	}

	fmt.Print("******************************************")
	pp.Print(caSecret)

	return nil
}

func createCertsMap(spec claiov1alpha1.ControlPlaneSpec) ([]Certificate, error) {
	certs := make([]Certificate, 0)

	// CA
	ca, err := CreateCaCert(nil, nil, nil)
	if err != nil {
		return nil, err
	}
	certs = append(certs, *ca)
	return certs, nil

	// kube-apiserver
	cert, err := CreateKubeApiServerCert(ca, []string{spec.AdvertiseHost}, []string{spec.AdvertiseAddress})
	if err != nil {
		return nil, err
	}
	certs = append(certs, *cert)

	// front-proxy-ca
	frontProxyCa, err := CreateFrontProxyCaCert()
	if err != nil {
		return nil, err
	}
	certs = append(certs, *frontProxyCa)

	// front-proxy-client
	cert, err = CreateFrontProxyClientCert(frontProxyCa)
	if err != nil {
		return nil, err
	}
	certs = append(certs, *cert)

	// apiserver-kubelet-client
	cert, err = CreateApiServerKubeletClientCert(ca)
	if err != nil {
		return nil, err
	}
	certs = append(certs, *cert)

	printCert(cert)

	return certs, nil
}

func printCert(cert *Certificate) {
	pcert, _ := x509.ParseCertificate(cert.RawCert)
	result, _ := certinfo.CertificateText(pcert)
	fmt.Print(result)
	fmt.Println(cert.PEM.Cert)

}
