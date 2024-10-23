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

func (k *KubeconfigFactory) GetAdminKubeconfig(namespace string) (string, error) {
	secretData, err := k.ControlPlaneSecretsFactory.k8s.GetSecret(namespace, "kubeconfig-admin")
	if err != nil {
		return "", fmt.Errorf("error getting kubeconfig secret %s/kubeconfig-admin: %s", namespace, err)
	}
	if secretData != nil {
		return string(secretData["super-admin.conf"]), nil
	}
	k.ControlPlaneSecretsFactory.log.Info("   create kubeconfig-admin")
	caCert, err := k.ControlPlaneSecretsFactory.GetCaCert(namespace)
	if err != nil {
		return "", fmt.Errorf("error getting ca cert in ns %s: %s", namespace, err)
	}
	clientCert, err := NewKubernetesAdminCert(namespace, "kubeconfig-admin", caCert, nil, nil)
	if err != nil {
		return "", fmt.Errorf("error creating kubconfig-admin certs in ns %s: %s", namespace, err)
	}
	kubeconfig := NewKubeconfig(
		namespace,
		k.ControlPlaneSecretsFactory.res.Spec.AdvertiseHost,
		"kubernetes-admin",
		caCert.Cert,
		clientCert.Cert,
		clientCert.Key,
	)
	yaml, err := kubeconfig.ToYaml()
	if err != nil {
		return "", fmt.Errorf("error converting kubeconfig-admin to yaml: %s", err)
	}
	if err := k.ControlPlaneSecretsFactory.k8s.CreateSecret(namespace, "kubeconfig-admin", map[string][]byte{"super-admin.conf": []byte(yaml)}); err != nil {
		return "", fmt.Errorf("error creating kubeconfig-admin secret in ns %s: %s", namespace, err)
	}
	return yaml, nil
}

func (k *KubeconfigFactory) GetSchedulerKubeconfig(namespace string) (string, error) {
	secretData, err := k.ControlPlaneSecretsFactory.k8s.GetSecret(namespace, "kubeconfig-scheduler")
	if err != nil {
		return "", fmt.Errorf("error getting kubeconfig secret %s/kubeconfig-scheduler: %s", namespace, err)
	}
	if secretData != nil {
		return string(secretData["scheduler.conf"]), nil
	}
	k.ControlPlaneSecretsFactory.log.Info("   create kubeconfig-scheduler")
	caCert, err := k.ControlPlaneSecretsFactory.GetCaCert(namespace)
	if err != nil {
		return "", fmt.Errorf("error getting ca cert in ns %s: %s", namespace, err)
	}
	clientCert, err := NewKubernetesSchedulerCert(namespace, "kubeconfig-scheduler", caCert, nil, nil)
	if err != nil {
		return "", fmt.Errorf("error creating kubconfig-scheduler certs in ns %s: %s", namespace, err)
	}
	kubeconfig := NewKubeconfig(
		"kubernetes",
		k.ControlPlaneSecretsFactory.res.Spec.AdvertiseHost,
		"system:kube-scheduler",
		caCert.Cert,
		clientCert.Cert,
		clientCert.Key,
	)
	yaml, err := kubeconfig.ToYaml()
	if err != nil {
		return "", fmt.Errorf("error converting kubeconfig-scheduler to yaml: %s", err)
	}
	if err := k.ControlPlaneSecretsFactory.k8s.CreateSecret(namespace, "kubeconfig-scheduler", map[string][]byte{"scheduler.conf": []byte(yaml)}); err != nil {
		return "", fmt.Errorf("error creating kubeconfig-scheduler secret in ns %s: %s", namespace, err)
	}
	return yaml, nil
}

func (k *KubeconfigFactory) GetControllerKubeconfig(namespace string) (string, error) {
	secretData, err := k.ControlPlaneSecretsFactory.k8s.GetSecret(namespace, "kubeconfig-controller")
	if err != nil {
		return "", fmt.Errorf("error getting kubeconfig secret %s/kubeconfig-controller: %s", namespace, err)
	}
	if secretData != nil {
		return string(secretData["controller-manager.conf"]), nil
	}
	k.ControlPlaneSecretsFactory.log.Info("   create kubeconfig-controller")
	caCert, err := k.ControlPlaneSecretsFactory.GetCaCert(namespace)
	if err != nil {
		return "", fmt.Errorf("error getting ca cert in ns %s: %s", namespace, err)
	}
	clientCert, err := NewKubernetesControllerCert(namespace, "kubeconfig-controller", caCert, nil, nil)
	if err != nil {
		return "", fmt.Errorf("error creating kubconfig-controller certs in ns %s: %s", namespace, err)
	}
	kubeconfig := NewKubeconfig(
		"kubernetes",
		k.ControlPlaneSecretsFactory.res.Spec.AdvertiseHost,
		"system:kube-controller-manager",
		caCert.Cert,
		clientCert.Cert,
		clientCert.Key,
	)
	yaml, err := kubeconfig.ToYaml()
	if err != nil {
		return "", fmt.Errorf("error converting kubeconfig-controller to yaml: %s", err)
	}
	if err := k.ControlPlaneSecretsFactory.k8s.CreateSecret(namespace, "kubeconfig-controller", map[string][]byte{"controller-manager.conf": []byte(yaml)}); err != nil {
		return "", fmt.Errorf("error creating kubeconfig-controller secret in ns %s: %s", namespace, err)
	}
	return yaml, nil
}

func (k *KubeconfigFactory) GetKonnectivityKubeconfig(namespace string) (string, error) {
	secretData, err := k.ControlPlaneSecretsFactory.k8s.GetSecret(namespace, "kubeconfig-konnectivity")
	if err != nil {
		return "", fmt.Errorf("error getting kubeconfig secret %s/kubeconfig-konnectivity: %s", namespace, err)
	}
	if secretData != nil {
		return string(secretData["konnectivity-server.conf"]), nil
	}
	k.ControlPlaneSecretsFactory.log.Info("   create kubeconfig-konnectivity")
	caCert, err := k.ControlPlaneSecretsFactory.GetCaCert(namespace)
	if err != nil {
		return "", fmt.Errorf("error getting ca cert in ns %s: %s", namespace, err)
	}
	clientCert, err := NewKubernetesControllerCert(namespace, "kubeconfig-konnectivity", caCert, nil, nil)
	if err != nil {
		return "", fmt.Errorf("error creating kubconfig-konnectivity certs in ns %s: %s", namespace, err)
	}
	kubeconfig := NewKubeconfig(
		"kubernetes",
		k.ControlPlaneSecretsFactory.res.Spec.AdvertiseHost,
		"system:konnectivity-server",
		caCert.Cert,
		clientCert.Cert,
		clientCert.Key,
	)
	yaml, err := kubeconfig.ToYaml()
	if err != nil {
		return "", fmt.Errorf("error converting kubeconfig-konnectivity to yaml: %s", err)
	}
	if err := k.ControlPlaneSecretsFactory.k8s.CreateSecret(namespace, "kubeconfig-konnectivity", map[string][]byte{"konnectivity-server.conf": []byte(yaml)}); err != nil {
		return "", fmt.Errorf("error creating kubeconfig-konnectivity secret in ns %s: %s", namespace, err)
	}
	return yaml, nil
}
