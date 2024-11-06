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

package kubeconfigs

import "fmt"

func (k *Factory) Check(caChanged bool) error {
	log := k.Factory.Base.Logger(1)
	log.Header("check kubeconfigs ...")
	// kubeconfig-admin
	_, _, err := k.GetAdminKubeconfig(caChanged)
	if err != nil {
		return fmt.Errorf("failed to get kubeconfig-admin")
	}
	// kubeconfig-scheduler
	_, _, err = k.GetSchedulerKubeconfig(caChanged)
	if err != nil {
		return fmt.Errorf("failed to get kubeconfig-scheduler")
	}
	// kubeconfig-controller
	_, _, err = k.GetControllerKubeconfig(caChanged)
	if err != nil {
		return fmt.Errorf("failed to get kubeconfig-controller")
	}
	// kubeconfig-konnectivity
	_, _, err = k.GetKonnectivityKubeconfig(caChanged)
	if err != nil {
		return fmt.Errorf("failed to get kubeconfig-konnectivity")
	}

	return nil
}
