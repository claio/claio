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

package k8s

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type K8s struct {
	ctx      context.Context
	client   client.Client
	scheme   *runtime.Scheme
	resource metav1.Object
}

func NewK8s(ctx context.Context, client client.Client, resource metav1.Object, scheme *runtime.Scheme) *K8s {
	return &K8s{
		ctx:      ctx,
		client:   client,
		resource: resource,
		scheme:   scheme,
	}
}

func (k *K8s) GetSecret(namespace string, name string) (map[string][]byte, error) {
	secret := &corev1.Secret{}
	if err := k.client.Get(
		k.ctx,
		k8sclient.ObjectKey{
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

func (k *K8s) CreateSecret(namespace string, name string, data map[string][]byte) error {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: data,
	}
	if err := ctrl.SetControllerReference(k.resource, secret, k.scheme); err != nil {
		return fmt.Errorf("   cannot set owner-reference on secret %s/%s: %s", namespace, name, err)
	}
	if err := k.client.Create(k.ctx, secret); err != nil {
		return fmt.Errorf("  failed to create secret %s/%s: %s", namespace, name, err)
	}
	return nil
}
