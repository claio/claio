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
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func GetDeployment(client k8sclient.Client, ctx context.Context, namespace, name string) (*appsv1.Deployment, error) {
	deployment := &appsv1.Deployment{}
	if err := client.Get(
		ctx,
		k8sclient.ObjectKey{
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

func CreateDeployment(client k8sclient.Client, ctx context.Context, namespace, name string, yaml []byte, reference client.Object, scheme *runtime.Scheme) error {
	decoder := serializer.NewCodecFactory(scheme).UniversalDecoder()
	deployment := &appsv1.Deployment{}
	if err := runtime.DecodeInto(decoder, yaml, deployment); err != nil {
		return fmt.Errorf("   cannot decode deployment %s/%s: %s", namespace, name, err)
	}
	if reference != nil {
		if err := ctrl.SetControllerReference(reference, deployment, scheme); err != nil {
			return fmt.Errorf("   cannot set owner-reference on deployment %s/%s: %s", namespace, name, err)
		}
	}
	if err := client.Create(ctx, deployment); err != nil {
		return fmt.Errorf("  failed to create deployment %s/%s: %s", namespace, name, err)
	}
	return nil
}

func UpdateDeployment(client k8sclient.Client, ctx context.Context, namespace, name string, yaml []byte, reference client.Object, scheme *runtime.Scheme) error {
	decoder := serializer.NewCodecFactory(scheme).UniversalDecoder()
	deployment := &appsv1.Deployment{}
	if err := runtime.DecodeInto(decoder, yaml, deployment); err != nil {
		return fmt.Errorf("   cannot decode deployment %s/%s: %s", namespace, name, err)
	}
	if reference != nil {
		if err := ctrl.SetControllerReference(reference, deployment, scheme); err != nil {
			return fmt.Errorf("   cannot set owner-reference on deployment %s/%s: %s", namespace, name, err)
		}
	}
	if err := client.Update(ctx, deployment); err != nil {
		return fmt.Errorf("  failed to create deployment %s/%s: %s", namespace, name, err)
	}
	return nil
}

func DeleteDeployment(client k8sclient.Client, ctx context.Context, namespace, name string, yaml []byte, reference client.Object, scheme *runtime.Scheme) error {
	decoder := serializer.NewCodecFactory(scheme).UniversalDecoder()
	deployment := &appsv1.Deployment{}
	if err := runtime.DecodeInto(decoder, yaml, deployment); err != nil {
		return fmt.Errorf("   cannot decode deployment %s/%s: %s", namespace, name, err)
	}
	if err := client.Delete(ctx, deployment); err != nil {
		return fmt.Errorf("  failed to create deployment %s/%s: %s", namespace, name, err)
	}
	return nil
}
