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

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type KubeconfigFactory struct {
	ControlPlaneSecretsFactory *ControlPlaneSecretsFactory
}

func NewKubeconfigFactory(client client.Client, res *claiov1alpha1.ControlPlane, ctx context.Context, scheme *runtime.Scheme, log *utils.Log) *KubeconfigFactory {
	return &KubeconfigFactory{
		ControlPlaneSecretsFactory: NewControlPlaneSecretsFactory(client, res, ctx, scheme, log),
	}
}

func NewKubeconfigFactoryFromSecretsFactory(controlPlaneSecretsFactory *ControlPlaneSecretsFactory) *KubeconfigFactory {
	return &KubeconfigFactory{
		ControlPlaneSecretsFactory: controlPlaneSecretsFactory,
	}
}

func (k *KubeconfigFactory) GetKubeconfig(namespace string) (string, error) {
	secretData, err := k.ControlPlaneSecretsFactory.k8s.GetSecret(namespace, "kubeconfig-admin")
	if err != nil {
		return "", fmt.Errorf("error getting kubeconfig secret %s/kubeconfig-admin: %s", namespace, err)
	}
	if secretData != nil {
		return string(secretData["admin"]), nil
	}
	k.ControlPlaneSecretsFactory.log.Info("   create kubeconfig-admin")
	caCert, err := k.ControlPlaneSecretsFactory.GetCaCert(namespace)
	if err != nil {
		return "", fmt.Errorf("error getting ca cert in ns %s: %s", namespace, err)
	}
	clientCert, err := NewKubernetesAdminCert(namespace, "kubeconfig-admin", caCert, nil, nil)
	if err != nil {
		return "", fmt.Errorf("error creating kubconfig certs in ns %s: %s", namespace, err)
	}
	kubeconfig := NewKubeconfig(
		namespace,
		k.ControlPlaneSecretsFactory.res.Spec.AdvertiseHost,
		caCert.Cert,
		clientCert.Cert,
		clientCert.Key,
	)
	yaml, err := kubeconfig.ToYaml()
	if err != nil {
		return "", fmt.Errorf("error converting kubeconfig to yaml: %s", err)
	}
	if err := k.ControlPlaneSecretsFactory.k8s.CreateSecret(namespace, "kubeconfig-admin", map[string][]byte{"admin.conf": []byte(yaml)}); err != nil {
		return "", fmt.Errorf("error creating kubeconfig secret in ns %s: %s", namespace, err)
	}
	return yaml, nil
}
