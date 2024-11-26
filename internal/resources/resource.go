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

package resources

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"strings"

	"claio/internal/kubernetes"

	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	finalizer = "claio.github.com/finalizer"
)

type Resource[T client.Object] struct {
	Scope            string
	Ctx              context.Context
	Req              ctrl.Request
	Client           client.Client
	Scheme           *runtime.Scheme
	Object           T
	STATUS_UP        string
	STATUS_WANTDOWN  string
	STATUS_GOINGDOWN string
}

func NewResource[T client.Object](scope string, ctx context.Context, req ctrl.Request, rClient client.Client, rScheme *runtime.Scheme, object T) *Resource[T] {
	return &Resource[T]{
		Scope:            scope,
		Ctx:              ctx,
		Req:              req,
		Client:           rClient,
		Scheme:           rScheme,
		Object:           object,
		STATUS_UP:        "UP",
		STATUS_WANTDOWN:  "WANTDOWN",
		STATUS_GOINGDOWN: "GOINGDOWN",
	}
}

// --- getter ---------------------------------------------------------------

func (r *Resource[T]) Namespace() string {
	return r.Req.Namespace
}

func (r *Resource[T]) Name() string {
	return r.Req.Name
}

// --- logging (protected) ----------------------------------------------------

func (r *Resource[T]) sprintf(tabs int, template string, args ...any) string {
	msg := fmt.Sprintf(template, args...)
	spaces := strings.Repeat("   ", tabs)
	return fmt.Sprintf("[%s]   %s%s", r.Namespace(), spaces, msg)
}

func (r *Resource[T]) LogHeader(template string, args ...any) {
	log.Log.WithName(r.Scope).Info(r.sprintf(0, template, args...))
}

func (r *Resource[T]) LogInfo(template string, args ...any) {
	log.Log.WithName(r.Scope).Info(r.sprintf(1, template, args...))
}

func (r *Resource[T]) LogError(err error, template string, args ...any) {
	log.Log.WithName(r.Scope).Error(err, r.sprintf(1, template, args...))
}

// --- kubernetes ------------------------------------------------------------

// finalizer

func (r *Resource[T]) HasFinalizer() bool {
	return controllerutil.ContainsFinalizer(r.Object, finalizer)
}

func (r *Resource[T]) AddFinalizer() error {
	controllerutil.AddFinalizer(r.Object, finalizer)
	if err := r.Client.Update(r.Ctx, r.Object); err != nil {
		return err
	}
	return nil
}

func (r *Resource[T]) RemoveFinalizer() error {
	controllerutil.RemoveFinalizer(r.Object, finalizer)
	if err := r.Client.Update(r.Ctx, r.Object); err != nil {
		return err
	}
	return nil
}

// secrets

func (r *Resource[T]) GetSecret(name string) (map[string][]byte, error) {
	return kubernetes.GetSecret(r.Client, r.Ctx, r.Namespace(), name)
}

func (r *Resource[T]) CreateSecret(name string, data map[string][]byte) error {
	return kubernetes.CreateSecret(r.Client, r.Ctx, r.Namespace(), name, data, r.Object, r.Scheme)
}

func (r *Resource[T]) DeleteSecret(name string) error {
	return kubernetes.DeleteSecret(r.Client, r.Ctx, r.Namespace(), name)
}

// deployments
func (r *Resource[T]) CreateDeployment(name string, yaml []byte) error {
	return kubernetes.CreateDeployment(r.Client, r.Ctx, r.Namespace(), name, yaml, r.Object, r.Scheme)
}

func (r *Resource[T]) UpdateDeployment(name string, yaml []byte) error {
	return kubernetes.UpdateDeployment(r.Client, r.Ctx, r.Namespace(), name, yaml, r.Object, r.Scheme)
}

func (r *Resource[T]) DeleteDeployment(name string, yaml []byte) error {
	return kubernetes.DeleteDeployment(r.Client, r.Ctx, r.Namespace(), name, yaml, r.Object, r.Scheme)
}

func (r *Resource[T]) GetDeployment(name string) (*v1.Deployment, error) {
	return kubernetes.GetDeployment(r.Client, r.Ctx, r.Namespace(), name)
}

// services
func (r *Resource[T]) CreateService(name string, yaml []byte) error {
	return kubernetes.CreateService(r.Client, r.Ctx, r.Namespace(), name, yaml, r.Object, r.Scheme)
}

func (r *Resource[T]) UpdateService(name string, yaml []byte) error {
	return kubernetes.UpdateService(r.Client, r.Ctx, r.Namespace(), name, yaml, r.Object, r.Scheme)
}

func (r *Resource[T]) DeleteService(name string, yaml []byte) error {
	return kubernetes.DeleteService(r.Client, r.Ctx, r.Namespace(), name, yaml, r.Object, r.Scheme)
}

func (r *Resource[T]) GetService(name string) (*corev1.Service, error) {
	return kubernetes.GetService(r.Client, r.Ctx, r.Namespace(), name)
}

// --- helper ----------------------------------------------------------------

func (c *Resource[T]) ToYaml(tmpl string, obj interface{}) ([]byte, error) {
	t, err := template.New("claio").Parse(tmpl)
	if err != nil {
		return nil, fmt.Errorf("error parsing template: %s", err)
	}
	var buf bytes.Buffer
	if err = t.Execute(&buf, obj); err != nil {
		return nil, fmt.Errorf("error executing template: %s", err)
	}
	return buf.Bytes(), nil
}
