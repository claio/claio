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

package controlplanes

import (
	claiov1alpha1 "claio/api/v1alpha1"
	"claio/internal/factory"
	"claio/internal/factory/kubernetes"
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type Factory struct {
	Base             *factory.Factory
	Resource         *claiov1alpha1.ControlPlane
	Namespace        string
	Name             string
	KubernetesClient *kubernetes.KubernetesClient
	Log              *factory.Log
}

func NewFactory(ctx context.Context, req ctrl.Request, rClient client.Client, rScheme *runtime.Scheme, res *claiov1alpha1.ControlPlane) *Factory {
	factory := factory.NewFactory("ControlPlane", ctx, req)
	factory.KubernetesClient = kubernetes.NewKubernetesClient(ctx, rClient, *rScheme, res)
	f := &Factory{
		Base:             factory,
		Resource:         res,
		Namespace:        factory.Namespace(),
		Name:             factory.Name(),
		KubernetesClient: factory.KubernetesClient,
	}
	return f
}

func (f *Factory) RemoveFinalizer() error {
	controllerutil.RemoveFinalizer(f.Resource, finalizer)
	if err := f.Base.KubernetesClient.Client.Update(f.Base.Ctx, f.Resource); err != nil {
		return err
	}
	return nil
}
