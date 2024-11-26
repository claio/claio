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

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func GetService(client k8sclient.Client, ctx context.Context, namespace, name string) (*corev1.Service, error) {
	service := &corev1.Service{}
	if err := client.Get(
		ctx,
		k8sclient.ObjectKey{
			Namespace: namespace,
			Name:      name,
		},
		service,
	); err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return service, nil
}

func CreateService(client k8sclient.Client, ctx context.Context, namespace, name string, yaml []byte, reference client.Object, scheme *runtime.Scheme) error {
	decoder := serializer.NewCodecFactory(scheme).UniversalDecoder()
	service := &corev1.Service{}
	if err := runtime.DecodeInto(decoder, yaml, service); err != nil {
		return fmt.Errorf("   cannot decode service %s/%s: %s", namespace, name, err)
	}
	if reference != nil {
		if err := ctrl.SetControllerReference(reference, service, scheme); err != nil {
			return fmt.Errorf("   cannot set owner-reference on service %s/%s: %s", namespace, name, err)
		}
	}
	if err := client.Create(ctx, service); err != nil {
		return fmt.Errorf("  failed to create service %s/%s: %s", namespace, name, err)
	}
	return nil
}

func UpdateService(client k8sclient.Client, ctx context.Context, namespace, name string, yaml []byte, reference client.Object, scheme *runtime.Scheme) error {
	decoder := serializer.NewCodecFactory(scheme).UniversalDecoder()
	service := &corev1.Service{}
	if err := runtime.DecodeInto(decoder, yaml, service); err != nil {
		return fmt.Errorf("   cannot decode service %s/%s: %s", namespace, name, err)
	}
	if reference != nil {
		if err := ctrl.SetControllerReference(reference, service, scheme); err != nil {
			return fmt.Errorf("   cannot set owner-reference on service %s/%s: %s", namespace, name, err)
		}
	}
	if err := client.Update(ctx, service); err != nil {
		return fmt.Errorf("  failed to create service %s/%s: %s", namespace, name, err)
	}
	return nil
}

func DeleteService(client k8sclient.Client, ctx context.Context, namespace, name string, yaml []byte, reference client.Object, scheme *runtime.Scheme) error {
	decoder := serializer.NewCodecFactory(scheme).UniversalDecoder()
	service := &corev1.Service{}
	if err := runtime.DecodeInto(decoder, yaml, service); err != nil {
		return fmt.Errorf("   cannot decode service %s/%s: %s", namespace, name, err)
	}
	if err := client.Delete(ctx, service); err != nil {
		return fmt.Errorf("  failed to create service %s/%s: %s", namespace, name, err)
	}
	return nil
}
