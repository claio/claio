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

package kubeconfigs

import (
	"claio/internal/factory"
	"claio/internal/resources/certificates"
	"fmt"
)

type KubeconfigFactory struct {
	Factory *factory.ControlPlaneFactory
}

func NewKubeconfigFactory(f *factory.ControlPlaneFactory) *KubeconfigFactory {
	return &KubeconfigFactory{
		Factory: f,
	}
}

// --- private ----------------------------------------------------------------
func (k *KubeconfigFactory) getKubeconfig(secretName, secretKey, clusterName, username string, forceCreate bool) ([]byte, bool, error) {
	factory := k.Factory
	namespace := factory.Namespace
	kubernetesClient := factory.KubernetesClient
	log := factory.Base.Logger(1)

	secretData, err := kubernetesClient.GetSecret(namespace, secretName)
	if err != nil {
		return nil, false, fmt.Errorf("error getting kubeconfig secret %s/%s: %s", namespace, secretName, err)
	}
	if secretData != nil {
		if !forceCreate {
			return secretData[secretKey], false, nil
		}
		log.Info("delete old/invalid secret: %s", secretName)
		if err := kubernetesClient.DeleteSecret(namespace, secretName); err != nil {
			return nil, true, fmt.Errorf("  failed to delete invalid secret %s: %s", secretName, err)
		}
	}
	log.Info("create %s", secretName)
	certificateFactory := certificates.NewCertificateFactory(factory)
	ca, err := certificateFactory.GetCertificateSecret("ca")
	if err != nil {
		return nil, true, fmt.Errorf("error getting ca cert in ns %s: %s", namespace, err)
	}
	clientCert, err := certificates.NewKubernetesAdminCert(ca, nil, nil)
	if err != nil {
		return nil, true, fmt.Errorf("error creating %s certs in ns %s: %s", secretName, namespace, err)
	}
	kubeconfig := NewKubeconfig(
		clusterName,
		factory.Spec.AdvertiseHost,
		username,
		ca.Cert,
		clientCert.Cert,
		clientCert.Key,
	)
	yaml, err := factory.Base.ToYaml(kubeconfigTemplate, kubeconfig)
	if err != nil {
		return nil, true, fmt.Errorf("error converting %s to yaml: %s", secretName, err)
	}
	if err := kubernetesClient.CreateSecret(namespace, secretName, map[string][]byte{secretKey: []byte(yaml)}); err != nil {
		return nil, true, fmt.Errorf("error creating %s secret in ns %s: %s", secretName, namespace, err)
	}
	return yaml, true, nil
}

// ----------------------------------------------------------------------------

func (k *KubeconfigFactory) GetAdminKubeconfig(forceCreate bool) ([]byte, bool, error) {
	return k.getKubeconfig("kubeconfig-admin", "super-admin.conf", k.Factory.Namespace, "kubernetes-admin", forceCreate)
}

func (k *KubeconfigFactory) GetSchedulerKubeconfig(forceCreate bool) ([]byte, bool, error) {
	return k.getKubeconfig("kubeconfig-scheduler", "scheduler.conf", "kubernetes", "system:kube-scheduler", forceCreate)
}

func (k *KubeconfigFactory) GetControllerKubeconfig(forceCreate bool) ([]byte, bool, error) {
	return k.getKubeconfig("kubeconfig-controller", "controller-manager.conf", "kubernetes", "system:kube-controller-manager", forceCreate)
}

func (k *KubeconfigFactory) GetKonnectivityKubeconfig(forceCreate bool) ([]byte, bool, error) {
	return k.getKubeconfig("kubeconfig-konnectivity", "konnectivity-server.conf", "kubernetes", "system:konnectivity-server", forceCreate)
}
