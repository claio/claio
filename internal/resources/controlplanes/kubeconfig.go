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
	b64 "encoding/base64"
	"fmt"
)

type Kubeconfig struct {
	ClusterName    string
	Server         string
	User           string
	CACertData     string
	ClientCertData string
	ClientKeyData  string
}

func NewKubeconfig(clusterName, server, user, caCertData, clientCertData, clientKeyData string) *Kubeconfig {
	return &Kubeconfig{
		ClusterName:    clusterName,
		Server:         server,
		User:           user,
		CACertData:     b64.StdEncoding.EncodeToString([]byte(caCertData)),
		ClientCertData: b64.StdEncoding.EncodeToString([]byte(clientCertData)),
		ClientKeyData:  b64.StdEncoding.EncodeToString([]byte(clientKeyData)),
	}
}

// --- private ----------------------------------------------------------------
func (c *ControlPlane) getKubeconfig(secretName, secretKey, clusterName, username string, forceCreate bool) ([]byte, bool, error) {
	secretData, err := c.GetSecret(secretName)
	if err != nil {
		return nil, false, fmt.Errorf("error getting kubeconfig secret %s/%s: %s", c.Namespace(), secretName, err)
	}
	if secretData != nil {
		if !forceCreate {
			return secretData[secretKey], false, nil
		}
		c.LogInfo("delete old/invalid secret: %s", secretName)
		if err := c.DeleteSecret(secretName); err != nil {
			return nil, true, fmt.Errorf("  failed to delete invalid secret %s: %s", secretName, err)
		}
	}
	c.LogInfo("create %s", secretName)
	ca, err := c.getCertificateSecret("ca")
	if err != nil {
		return nil, true, fmt.Errorf("error getting ca cert in ns %s: %s", c.Namespace(), err)
	}
	clientCert, err := newKubernetesAdminCert(ca, nil, nil)
	if err != nil {
		return nil, true, fmt.Errorf("error creating %s certs in ns %s: %s", secretName, c.Namespace(), err)
	}
	kubeconfig := NewKubeconfig(
		clusterName,
		c.Object.Spec.AdvertiseHost,
		username,
		ca.Cert,
		clientCert.Cert,
		clientCert.Key,
	)
	yaml, err := c.ToYaml(kubeconfigTemplate, kubeconfig)
	if err != nil {
		return nil, true, fmt.Errorf("error converting %s to yaml: %s", secretName, err)
	}
	if err := c.CreateSecret(secretName, map[string][]byte{secretKey: []byte(yaml)}); err != nil {
		return nil, true, fmt.Errorf("error creating %s secret in ns %s: %s", secretName, c.Namespace(), err)
	}
	return yaml, true, nil
}

// ----------------------------------------------------------------------------

func (c *ControlPlane) GetAdminKubeconfig(forceCreate bool) ([]byte, bool, error) {
	return c.getKubeconfig("kubeconfig-admin", "super-admin.conf", c.Namespace(), "kubernetes-admin", forceCreate)
}

func (c *ControlPlane) GetSchedulerKubeconfig(forceCreate bool) ([]byte, bool, error) {
	return c.getKubeconfig("kubeconfig-scheduler", "scheduler.conf", "kubernetes", "system:kube-scheduler", forceCreate)
}

func (c *ControlPlane) GetControllerKubeconfig(forceCreate bool) ([]byte, bool, error) {
	return c.getKubeconfig("kubeconfig-controller", "controller-manager.conf", "kubernetes", "system:kube-controller-manager", forceCreate)
}

func (c *ControlPlane) GetKonnectivityKubeconfig(forceCreate bool) ([]byte, bool, error) {
	return c.getKubeconfig("kubeconfig-konnectivity", "konnectivity-server.conf", "kubernetes", "system:konnectivity-server", forceCreate)
}

func (c *ControlPlane) kubeconfigReconcile(caChanged bool) error {
	c.LogHeader("check kubeconfigs ...")
	// kubeconfig-admin
	_, _, err := c.GetAdminKubeconfig(caChanged)
	if err != nil {
		return fmt.Errorf("failed to get kubeconfig-admin")
	}
	// kubeconfig-scheduler
	_, _, err = c.GetSchedulerKubeconfig(caChanged)
	if err != nil {
		return fmt.Errorf("failed to get kubeconfig-scheduler")
	}
	// kubeconfig-controller
	_, _, err = c.GetControllerKubeconfig(caChanged)
	if err != nil {
		return fmt.Errorf("failed to get kubeconfig-controller")
	}
	// kubeconfig-konnectivity
	_, _, err = c.GetKonnectivityKubeconfig(caChanged)
	if err != nil {
		return fmt.Errorf("failed to get kubeconfig-konnectivity")
	}

	return nil
}

const kubeconfigTemplate = `
apiVersion: v1
kind: Config
clusters:
  - name: {{ .ClusterName }}
    cluster:
      certificate-authority-data: {{ .CACertData }}
      server: "https://{{ .Server }}:6543"
contexts:
  - name: {{ .User }}@{{ .ClusterName }}
    context:		
      cluster: {{ .ClusterName }}
      user: {{ .User }}
current-context: {{ .User }}@{{ .ClusterName }}
preferences: {}
users:
  - name: {{ .User }}
    user:
      client-certificate-data: {{ .ClientCertData }}
      client-key-data: {{ .ClientKeyData }}
`
