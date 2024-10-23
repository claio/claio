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

package certificates

import (
	"bytes"
	b64 "encoding/base64"
	"fmt"
	"text/template"
)

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

func (k *Kubeconfig) ToYaml() (string, error) {
	tmpl, err := template.New("kubeconfig-admin").Parse(kubeconfigTemplate)
	if err != nil {
		return "", fmt.Errorf("error parsing template: %s", err)
	}
	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, k); err != nil {
		return "", fmt.Errorf("error executing template: %s", err)
	}
	return buf.String(), nil
}
