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

package deployments

import (
	"reflect"
	"time"
)

func (c *ControlPlaneDeploymentFactory) Check() error {
	log := c.Factory.Base.Logger(1)
	log.Info("   check control-plane deployment ...")
	deployment, err := c.GetDeployment(c.Factory.Namespace, "claio")
	if err != nil {
		log.Error(err, "failed to retreive claio deployment")
		return err
	}
	// deployment does not exist - create it
	if deployment == nil {
		log.Info("      create claio deployment")
		if err := c.CreateDeployment(c.Factory.Namespace, "claio"); err != nil {
			log.Error(err, "      failed to create deployment")
			return err
		}
		return nil
	}
	// deployment exists - update it if controlpane resource has changed
	if !c.isEqual() {
		if c.Factory.Spec.Database != c.Factory.Status.TargetSpec.Database {
			log.Info("      database changed - need to stop control-plane")
			if err := c.DeleteDeployment(c.Factory.Namespace, "claio"); err != nil {
				log.Error(err, "       failed to delete deployment")
				return err
			}
			log.Info("      deployment stopped")
			log.Info("      migrating database ... (fake) ... (sleep 10)")
			time.Sleep(10 * time.Second)
			log.Info("      database migration done")
			// the deployment will be startet with the next reconcilation run
			return nil
		}
		log.Info("      update claio deployment")
		if err := c.UpdateDeployment(c.Factory.Namespace, "claio"); err != nil {
			log.Error(err, "       failed to update deployment")
			return err
		}
		return nil
	}

	return nil
}

func (c *ControlPlaneDeploymentFactory) isEqual() bool {
	return reflect.DeepEqual(*c.Factory.Spec, c.Factory.Status.TargetSpec)
}
