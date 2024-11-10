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
	"claio/internal/resources"
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ControlPlane struct {
	resources.Resource[*claiov1alpha1.ControlPlane]
}

func NewControlPlane(ctx context.Context, req ctrl.Request, rClient client.Client, rScheme *runtime.Scheme) (*ControlPlane, error) {
	res := &claiov1alpha1.ControlPlane{}
	if err := rClient.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, res); err != nil {
		return nil, client.IgnoreNotFound(err)
	}
	return &ControlPlane{
		Resource: *resources.NewResource("ControlPlane", ctx, req, rClient, rScheme, res),
	}, nil
}

func (r *ControlPlane) Check() (string, error) {
	r.LogHeader("check control-plane (init) ...")
	if r.Object.ObjectMeta.DeletionTimestamp.IsZero() {
		if !r.HasFinalizer() {
			r.LogInfo("add finalizer")
			if err := r.AddFinalizer(); err != nil {
				return r.STATUS_UP, fmt.Errorf("adding finalizer failed")
			}
		}
		return r.STATUS_UP, nil
	} else {
		if r.HasFinalizer() {
			return r.STATUS_WANTDOWN, nil
		}
		return r.STATUS_GOINGDOWN, nil
	}
}

func (r *ControlPlane) Reconcile() error {
	status, err := r.Check()
	if err != nil {
		r.LogError(err, "check failed")
		return err
	}
	r.LogInfo("status: %s", status)

	apiDirty := false
	if status == r.STATUS_UP {
		// check certificates
		caChanged, localApiDirty, err := r.reconcileCertificates()
		if err != nil {
			r.LogError(err, "failed to reconcile secrets")
			return err
		}
		apiDirty = apiDirty || localApiDirty

		if err := r.kubeconfigReconcile(caChanged); err != nil {
			r.LogError(err, "failed to reconcile kubeconfig")
			return err
		}
	}

	// check deployment
	if status == r.STATUS_UP || status == r.STATUS_WANTDOWN {
		if err := r.ReconcileDeployment(apiDirty, status); err != nil {
			r.LogError(err, "failed to check deployment")
			return err
		}
	}

	// handle finalizer
	r.LogHeader("check control-plane (finalize) ...")
	if status == r.STATUS_WANTDOWN {
		r.LogInfo("remove finalizer")
		if err := r.RemoveFinalizer(); err != nil {
			r.LogError(err, "failed to remove finalizer")
			return err
		}
		return nil
	}

	// update status
	r.Object.Status = claiov1alpha1.ControlPlaneStatus{TargetSpec: r.Object.Spec}
	if err := r.Client.Status().Update(r.Ctx, r.Object); err != nil {
		r.LogError(err, "failed to update status")
		return err
	}

	return nil
}
