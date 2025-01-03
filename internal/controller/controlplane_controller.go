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

package controller

import (
	"context"
	"reflect"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	claiov1alpha1 "claio/api/v1alpha1"
	"claio/internal/resources/controlplanes"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// ControlPlaneReconciler reconciles a ControlPlane object
type ControlPlaneReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=claio.github.com,resources=controlplanes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=claio.github.com,resources=controlplanes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=claio.github.com,resources=controlplanes/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="apps",resources=deployments,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ControlPlane object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.2/pkg/reconcile
func (r *ControlPlaneReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	controlPlane, err := controlplanes.NewControlPlane(ctx, req, r.Client, r.Scheme)
	if err != nil {
		return ctrl.Result{}, err
	}
	if controlPlane == nil {
		return ctrl.Result{}, nil
	}
	controlPlane.LogHeader("--- Reconciling --------------------------------------")
	err = controlPlane.Reconcile()
	controlPlane.LogHeader("--- Reconciling Done ---------------------------------")
	return ctrl.Result{}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *ControlPlaneReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&claiov1alpha1.ControlPlane{}).
		Owns(&corev1.Secret{}).
		Owns(&corev1.Service{}).
		Owns(&appsv1.Deployment{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 1,
		}).
		WithEventFilter(predicate.Funcs{
			CreateFunc: func(e event.CreateEvent) bool {
				return reflect.TypeOf(e.Object) == reflect.TypeOf(&claiov1alpha1.ControlPlane{})
			},
			UpdateFunc: func(e event.UpdateEvent) bool {
				if reflect.TypeOf(e.ObjectNew) == reflect.TypeOf(&claiov1alpha1.ControlPlane{}) {
					if e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration() {
						return true
					}
				}
				return false
			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				isSecret := reflect.TypeOf(e.Object) == reflect.TypeOf(&corev1.Secret{})
				isService := reflect.TypeOf(e.Object) == reflect.TypeOf(&corev1.Service{})
				isDeployment := reflect.TypeOf(e.Object) == reflect.TypeOf(&appsv1.Deployment{})
				return isSecret || isDeployment || isService
			},
		}).
		Complete(r)
}
