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

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	claiov1alpha1 "claio/api/v1alpha1"

	"claio/internal/factory"
	"claio/internal/resources/certificates"
	"claio/internal/resources/deployments"
	"claio/internal/resources/kubeconfigs"

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
	log := factory.NewLog("ControlPlane", req.Namespace, req.Name)
	log.Info("--- Reconciling --------------------------------------")

	// fetch the ControlPlane instance
	res := &claiov1alpha1.ControlPlane{}
	if err := r.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, res); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	factory := factory.NewControlPlaneFactory(ctx, req, r.Client, r.Scheme, res)

	// check secrets
	certicateFactory := certificates.NewCertificateFactory(factory)
	if err := certicateFactory.Check(); err != nil {
		log.Error(err, "failed to check secrets")
	}
	kubeconfigFactory := kubeconfigs.NewKubeconfigFactory(factory)
	if err := kubeconfigFactory.Check(); err != nil {
		log.Error(err, "failed to check kubeconfigs")
	}

	// check deployment
	deploymentFactory := deployments.NewControlPlaneDeploymentFactory(factory)
	if err := deploymentFactory.Check(); err != nil {
		log.Error(err, "failed to check deployment")
	}

	log.Info("Reconciling done")

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ControlPlaneReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&claiov1alpha1.ControlPlane{}).
		Owns(&corev1.Secret{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
