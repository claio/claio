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

	"claio/internal/certificates"
	"claio/internal/utils"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	log := utils.NewLog("ControlPlane", req.Namespace, req.Name)

	// fetch the ControlPlane instance
	res := &claiov1alpha1.ControlPlane{}
	r.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, res)

	// check secrets (create them if necessary)
	err := r.checkSecrets(req.Namespace, res.Spec, ctx, log)
	if err != nil {
		log.Error(err, "failed to create secrets")
	}
	log.Info("secrets done")

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ControlPlaneReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&claiov1alpha1.ControlPlane{}).
		Complete(r)
}

// ---------------------------------------------------
func (r *ControlPlaneReconciler) getSecret(namespace string, name string, ctx context.Context) (*corev1.Secret, error) {
	secret := &corev1.Secret{}
	if err := r.Get(
		ctx,
		client.ObjectKey{
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
	return secret, nil
}

func (r *ControlPlaneReconciler) checkSecret(namespace string, name string, ca *certificates.CertificatePEM, fn certificates.CertificateCreator, spec claiov1alpha1.ControlPlaneSpec, ctx context.Context, log *utils.Log) (pem *certificates.CertificatePEM, err error) {
	secret, err := r.getSecret(namespace, name, ctx)
	if err != nil {
		return nil, err
	}
	if secret != nil {
		return &certificates.CertificatePEM{
			Cert: string(secret.Data[name+".crt"]),
			Key:  string(secret.Data[name+".key"]),
		}, nil
	}
	log.Info("   create certificate and secret: %s", name)
	cert, err := fn(ca, &spec.AdvertiseHost, &spec.AdvertiseAddress)
	if err != nil {
		return nil, err
	}
	// create secret
	data := make(map[string][]byte)
	data[name+".crt"] = []byte(cert.PEM.Cert)
	data[name+".key"] = []byte(cert.PEM.Key)
	secret = &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: data,
	}
	if err := r.Create(ctx, secret); err != nil {
		return nil, err
	}
	return cert.PEM, nil
}

func (r *ControlPlaneReconciler) checkSecrets(namespace string, spec claiov1alpha1.ControlPlaneSpec, ctx context.Context, log *utils.Log) error {
	log.Info("Check secrets ...")
	_, err := r.checkSecret(namespace, "ca", nil, certificates.CreateCaCert, spec, ctx, log)
	if err != nil {
		return err
	}
	return nil
}
