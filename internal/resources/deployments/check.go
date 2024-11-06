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
	"claio/internal/resources/controlplanes"
	"fmt"
	"reflect"
	"time"
)

func (c *Factory) Check(apiDirty bool, mode int) error {
	log := c.Factory.Base.Logger(1)
	log.Header("check control-plane deployment ...")
	deployment, err := c.GetDeployment(c.Factory.Namespace, "claio")
	if err != nil {
		log.Error(err, "failed to retreive claio deployment")
		return err
	}

	// stop deployment, controlplane wants to stop
	if mode == controlplanes.WANTDOWN {
		if deployment == nil {
			log.Info("deployment already deleted")
		} else {
			if err := c.stopDeployment(); err != nil {
				return err
			}
		}
		return nil
	}

	if deployment == nil {
		log.Info("create claio deployment")
		if err := c.CreateDeployment(c.Factory.Namespace, "claio"); err != nil {
			log.Error(err, "failed to create deployment")
			return err
		}
		return nil
	}

	if apiDirty || !c.isEqual() {
		log.Info("structural changes detected - need to stop control-plane")
		if err := c.stopDeployment(); err != nil {
			log.Error(err, "failed to stop deployment")
			return err
		}
		/*
			if c.Factory.Resource.Spec.Database != c.Factory.Resource.Status.TargetSpec.Database {
				log.Info("migrating database ... (fake) ... (sleep 10)")
				time.Sleep(10 * time.Second)
				log.Info("database migration done")
			}
		*/

		// the deployment will be startet with the next reconcilation run
		return nil
	}
	return nil
}

func (c *Factory) isEqual() bool {
	return reflect.DeepEqual(c.Factory.Resource.Spec, c.Factory.Resource.Status.TargetSpec)
}

func (c *Factory) stopDeployment() error {
	log := c.Factory.Base.Logger(1)
	log.Info("stop deployment")
	if err := c.DeleteDeployment(c.Factory.Namespace, "claio"); err != nil {
		log.Error(err, "failed to delete deployment")
		return err
	}
	loop := 0
	for {
		if loop > 10 {
			return fmt.Errorf("failed to stop deployment")
		}
		depl, err := c.GetDeployment(c.Factory.Namespace, "claio")
		if err == nil && depl == nil {
			break
		}
		time.Sleep(3 * time.Second)
		loop++
	}
	log.Info("deployment stopped")
	return nil
}
