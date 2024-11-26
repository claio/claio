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
	"fmt"
	"reflect"

	v1 "k8s.io/api/core/v1"
)

func (c *ControlPlane) CreateClaioService() error {
	yaml, err := c.ToYaml(controlplaneServiceTemplate, c.Object.Spec)
	if err != nil {
		return fmt.Errorf("error generating yaml: %s", err)
	}
	return c.CreateService("claio", yaml)
}

func (c *ControlPlane) UpdateClaioService() error {
	yaml, err := c.ToYaml(controlplaneServiceTemplate, c.Object.Spec)
	if err != nil {
		return fmt.Errorf("error generating yaml: %s", err)
	}
	return c.UpdateDeployment("claio", yaml)
}

func (c *ControlPlane) DeleteClaioService() error {
	yaml, err := c.ToYaml(controlplaneServiceTemplate, c.Object.Spec)
	if err != nil {
		return fmt.Errorf("error generating yaml: %s", err)
	}
	return c.DeleteService("claio", yaml)
}

func (c *ControlPlane) GetClaioService() (*v1.Service, error) {
	service, err := c.GetService("claio-apiserver")
	if err != nil {
		return nil, fmt.Errorf("error getting service: %s", err)
	}
	return service, nil
}

func (c *ControlPlane) ReconcileService(apiDirty bool, mode string) error {
	c.LogHeader("check service ...")
	service, err := c.GetClaioService()
	if err != nil {
		c.LogError(err, "failed to retreive claio service")
		return err
	}

	if mode != c.STATUS_WANTDOWN && service == nil {
		c.LogInfo("create claio service")
		if err := c.CreateClaioService(); err != nil {
			c.LogError(err, "failed to create service")
			return err
		}
		return nil
	}

	if apiDirty || !reflect.DeepEqual(c.Object.Spec, c.Object.Status.TargetSpec) {
		c.LogInfo("structural changes detected - need to recreate service")
		if err := c.DeleteClaioService(); err != nil {
			c.LogError(err, "failed to delete service")
			return err
		}

		// the service will be startet with the next reconcilation run
		return nil
	}

	return nil
}

const controlplaneServiceTemplate = `apiVersion: v1
kind: Service
metadata:
  name: claio-apiserver
  labels:
    app: claio-apiserver
  namespace: tenant-{{ .Name }}
spec:
  type: LoadBalancer
  selector:
    app: claio
  ports:
  - port: {{ .Port }}
    targetPort: {{ .Port }}
    protocol: TCP
    name: https
`
