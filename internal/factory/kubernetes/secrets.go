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

package kubernetes

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (k *KubernetesClient) GetSecret(namespace, name string) (map[string][]byte, error) {
	secret := &corev1.Secret{}
	if err := k.Client.Get(
		k.Ctx,
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
	return secret.Data, nil
}

func (k *KubernetesClient) CreateSecret(namespace, name string, data map[string][]byte) error {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: data,
	}
	if err := ctrl.SetControllerReference(k.Resource, secret, &k.Scheme); err != nil {
		return fmt.Errorf("   cannot set owner-reference on secret %s/%s: %s", namespace, name, err)
	}
	if err := k.Client.Create(k.Ctx, secret); err != nil {
		return fmt.Errorf("  failed to create secret %s/%s: %s", namespace, name, err)
	}
	return nil
}

func (k *KubernetesClient) DeleteSecret(namespace, name string) error {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	if err := k.Client.Delete(k.Ctx, secret); err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("  failed to delete secret %s/%s: %s", namespace, name, err)
	}
	return nil
}
