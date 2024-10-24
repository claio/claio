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

func (c *ControlPlaneDeploymentFactory) Check() error {
	log := c.Factory.Log
	log.Info("   check control-plane deployment ...")
	if deployment, err := c.GetDeployment(c.Factory.Namespace, "claio"); err == nil {
		if deployment == nil {
			log.Info("   create claio deployment")
			if err := c.CreateDeployment(c.Factory.Namespace, "claio"); err != nil {
				log.Error(err, "failed to reconcile control-plane")
				return err
			}
			return nil
		}
		return nil
	} else {
		log.Error(err, "failed to retreive claio deployment")
		return err
	}
}
