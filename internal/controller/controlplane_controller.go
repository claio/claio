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
	"fmt"

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
	log.Info("--- Reconciling --------------------------------------")

	// fetch the ControlPlane instance
	res := &claiov1alpha1.ControlPlane{}
	r.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, res)

	// check secrets (create them if necessary)
	err := r.checkSecrets(req.Namespace, res, ctx, log)
	if err != nil {
		log.Error(err, "failed to create secrets")
	}
	log.Info("Reconciling done")

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ControlPlaneReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&claiov1alpha1.ControlPlane{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}

// ---------------------------------------------------
func (r *ControlPlaneReconciler) checkSecrets(namespace string, res *claiov1alpha1.ControlPlane, ctx context.Context, log *utils.Log) error {
	log.Info("Check secrets ...")
	caCert, err := r.checkSecret(namespace, "ca", nil, certificates.NewCaCert, res, ctx, log)
	if err != nil {
		return err
	}
	if _, err := r.checkSecret(namespace, "apiserver", caCert, certificates.NewApiserverCert, res, ctx, log); err != nil {
		return err
	}
	if _, err := r.checkSecret(namespace, "apiserver-kubelet-client", caCert, certificates.NewApiserverKubeletClientCert, res, ctx, log); err != nil {
		return err
	}
	frontProxyCaCert, err := r.checkSecret(namespace, "front-proxy-ca", nil, certificates.NewFrontProxyCaCert, res, ctx, log)
	if err != nil {
		return err
	}
	if _, err := r.checkSecret(namespace, "front-proxy-client", frontProxyCaCert, certificates.NewFrontProxyClientCert, res, ctx, log); err != nil {
		return err
	}
	if _, err := r.checkSecret(namespace, "sa", nil, certificates.NewSaRSA, res, ctx, log); err != nil {
		return err
	}
	return nil
}

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

func (r *ControlPlaneReconciler) checkSecret(namespace string, name string, ca *certificates.Certificate, fn certificates.CertificateCreator, res *claiov1alpha1.ControlPlane, ctx context.Context, log *utils.Log) (pem *certificates.Certificate, err error) {
	secret, err := r.getSecret(namespace, name, ctx)
	if err != nil {
		return nil, fmt.Errorf("  failed to get secret %s/%s: %s", namespace, name, err)
	}
	if secret != nil {
		cert := certificates.Certificate{
			Name:      name,
			Namespace: namespace,
			Key:       string(secret.Data[name+".key"]),
			Cert:      "",
			Pub:       "",
		}
		if _, found := secret.Data[name+".crt"]; found {
			cert.Cert = string(secret.Data[name+".crt"])
		}
		if _, found := secret.Data[name+".pub"]; found {
			cert.Pub = string(secret.Data[name+".pub"])
		}
		return &cert, nil
	}
	log.Info("   create certificate and secret: %s", name)
	cert, err := fn(namespace, name, ca, &res.Spec.AdvertiseHost, &res.Spec.AdvertiseAddress)
	if err != nil {
		return nil, fmt.Errorf("  failed to get secret %s: %s", name, err)
	}
	// create secret
	data := make(map[string][]byte)
	data[name+".key"] = []byte(cert.Key)
	if cert.Cert != "" {
		data[name+".crt"] = []byte(cert.Cert)
	}
	if cert.Pub != "" {
		data[name+".pub"] = []byte(cert.Pub)
	}
	secret = &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: data,
	}
	if err := ctrl.SetControllerReference(res, secret, r.Scheme); err != nil {
		return nil, fmt.Errorf("   cannot set owner-reference on secret %s/%s: %s", namespace, name, err)
	}
	if err := r.Create(ctx, secret); err != nil {
		return nil, fmt.Errorf("  failed to create secret %s/%s: %s", namespace, name, err)
	}
	return cert, nil
}
