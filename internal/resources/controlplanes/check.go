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

import "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

const (
	finalizer = "claio.github.com/finalizer"
	UP        = iota
	WANTDOWN
	MIGRATE
)

func (c *Factory) Check() (int, error) {
	log := c.Base.Logger(1)
	log.Header("check control-plane ...")

	if c.Resource.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		if !controllerutil.ContainsFinalizer(c.Resource, finalizer) {
			log.Info("add finalizer")
			controllerutil.AddFinalizer(c.Resource, finalizer)
			if err := c.Base.KubernetesClient.Client.Update(c.Base.Ctx, c.Resource); err != nil {
				return UP, err
			}
		}
	} else {
		// object should be deleted
		if controllerutil.ContainsFinalizer(c.Resource, finalizer) {
			return WANTDOWN, nil
		}
	}
	return UP, nil
}
