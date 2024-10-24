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
	"context"

	"claio/internal/factory/kubernetes"

	ctrl "sigs.k8s.io/controller-runtime"
)

type Factory struct {
	Scope            string
	Ctx              context.Context
	Req              ctrl.Request
	Log              *Log
	KubernetesClient *kubernetes.KubernetesClient
}

type FactoryExtension struct {
	Log *Log
}

func NewFactory(scope string, ctx context.Context, req ctrl.Request) *Factory {
	return &Factory{
		Scope:            scope,
		Ctx:              ctx,
		Req:              req,
		Log:              NewLog(scope, req.Namespace, req.Name),
		KubernetesClient: nil,
	}
}

func (f *Factory) Namespace() string {
	return f.Req.Namespace
}

func (f *Factory) Name() string {
	return f.Req.Name
}
