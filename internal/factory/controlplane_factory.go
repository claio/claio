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

package factory

import (
	claiov1alpha1 "claio/api/v1alpha1"
	"claio/internal/factory/kubernetes"
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ControlPlaneFactory struct {
	Base             *Factory
	Spec             *claiov1alpha1.ControlPlaneSpec
	Namespace        string
	Name             string
	KubernetesClient *kubernetes.KubernetesClient
	Log              *Log
}

func NewControlPlaneFactory(ctx context.Context, req ctrl.Request, rClient client.Client, rScheme *runtime.Scheme, res *claiov1alpha1.ControlPlane) *ControlPlaneFactory {
	factory := NewFactory("ControlPlane", ctx, req)
	factory.KubernetesClient = kubernetes.NewKubernetesClient(ctx, rClient, *rScheme, res)
	f := &ControlPlaneFactory{
		Base:             factory,
		Spec:             &res.Spec,
		Namespace:        factory.Namespace(),
		Name:             factory.Name(),
		KubernetesClient: factory.KubernetesClient,
		Log:              factory.Log,
	}
	return f
}
