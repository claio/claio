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

	appsv1 "k8s.io/api/apps/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (k *KubernetesClient) GetDeployment(namespace, name string) (*appsv1.Deployment, error) {
	deployment := &appsv1.Deployment{}
	if err := k.Client.Get(
		k.Ctx,
		client.ObjectKey{
			Namespace: namespace,
			Name:      name,
		},
		deployment,
	); err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return deployment, nil
}

func (k *KubernetesClient) CreateDeployment(namespace, name string, yaml []byte) error {
	decoder := serializer.NewCodecFactory(&k.Scheme).UniversalDecoder()
	deployment := &appsv1.Deployment{}
	if err := runtime.DecodeInto(decoder, yaml, deployment); err != nil {
		return fmt.Errorf("   cannot decode deployment %s/%s: %s", namespace, name, err)
	}
	if err := ctrl.SetControllerReference(k.Resource, deployment, &k.Scheme); err != nil {
		return fmt.Errorf("   cannot set owner-reference on deployment %s/%s: %s", namespace, name, err)
	}
	if err := k.Client.Create(k.Ctx, deployment); err != nil {
		return fmt.Errorf("  failed to create deployment %s/%s: %s", namespace, name, err)
	}
	return nil
}

func (k *KubernetesClient) UpdateDeployment(namespace, name string, yaml []byte) error {
	decoder := serializer.NewCodecFactory(&k.Scheme).UniversalDecoder()
	deployment := &appsv1.Deployment{}
	if err := runtime.DecodeInto(decoder, yaml, deployment); err != nil {
		return fmt.Errorf("   cannot decode deployment %s/%s: %s", namespace, name, err)
	}
	if err := ctrl.SetControllerReference(k.Resource, deployment, &k.Scheme); err != nil {
		return fmt.Errorf("   cannot set owner-reference on deployment %s/%s: %s", namespace, name, err)
	}
	if err := k.Client.Update(k.Ctx, deployment); err != nil {
		return fmt.Errorf("  failed to create deployment %s/%s: %s", namespace, name, err)
	}
	return nil
}

func (k *KubernetesClient) DeleteDeployment(namespace, name string, yaml []byte) error {
	decoder := serializer.NewCodecFactory(&k.Scheme).UniversalDecoder()
	deployment := &appsv1.Deployment{}
	if err := runtime.DecodeInto(decoder, yaml, deployment); err != nil {
		return fmt.Errorf("   cannot decode deployment %s/%s: %s", namespace, name, err)
	}
	if err := k.Client.Delete(k.Ctx, deployment); err != nil {
		return fmt.Errorf("  failed to create deployment %s/%s: %s", namespace, name, err)
	}
	return nil
}
