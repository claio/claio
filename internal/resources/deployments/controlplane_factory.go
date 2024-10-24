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
	"claio/internal/factory"
	"fmt"

	v1 "k8s.io/api/apps/v1"
)

type ControlPlaneDeploymentFactory struct {
	Factory *factory.ControlPlaneFactory
}

func NewControlPlaneDeploymentFactory(factory *factory.ControlPlaneFactory) *ControlPlaneDeploymentFactory {
	return &ControlPlaneDeploymentFactory{
		Factory: factory,
	}
}

func (c *ControlPlaneDeploymentFactory) CreateDeployment(namespace, name string) error {
	deploymentYaml, err := c.Factory.Base.ToYaml(controlplaneTemplate, c.Factory.Spec)
	if err != nil {
		return fmt.Errorf("error generating yaml: %s", err)
	}
	return c.Factory.KubernetesClient.CreateDeployment(c.Factory.Namespace, c.Factory.Name, deploymentYaml)
}

func (c *ControlPlaneDeploymentFactory) GetDeployment(namespace, name string) (*v1.Deployment, error) {
	deployment, err := c.Factory.KubernetesClient.GetDeployment(namespace, name)
	if err != nil {
		return nil, fmt.Errorf("error getting deployment: %s", err)
	}
	return deployment, nil
}

const controlplaneTemplate = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: claio
  labels:
    app: claio
  namespace: tenant-{{ .Name }}
spec:
  replicas: 1
  selector:
    matchLabels:
      app: claio
  template:
    metadata:
      labels:
        app: claio
    spec:
      containers:
        - name: kube-apiserver
          image: registry.k8s.io/kube-apiserver:v{{ .Version }}
          command:
            - kube-apiserver
          args:      
            #- --advertise-address=127.0.0.1
            - --allow-privileged=true
            - --authorization-mode=Node,RBAC
            - --client-ca-file=/etc/kubernetes/pki/ca.crt
            - --enable-bootstrap-token-auth=true
            - --etcd-prefix=/{{ .Name }}
            - --etcd-servers={{ .Database }}
            - --external-hostname={{ .AdvertiseHost }}
            - --kubelet-client-certificate=/etc/kubernetes/pki/apiserver-kubelet-client.crt
            - --kubelet-client-key=/etc/kubernetes/pki/apiserver-kubelet-client.key
            - --kubelet-preferred-address-types=InternalIP,ExternalIP,Hostname
            - --proxy-client-cert-file=/etc/kubernetes/pki/front-proxy-client.crt
            - --proxy-client-key-file=/etc/kubernetes/pki/front-proxy-client.key
            - --requestheader-allowed-names=front-proxy-client
            - --requestheader-client-ca-file=/etc/kubernetes/pki/front-proxy-ca.crt
            - --requestheader-extra-headers-prefix=X-Remote-Extra-
            - --requestheader-group-headers=X-Remote-Group
            - --requestheader-username-headers=X-Remote-User
            - --secure-port={{ .Port }}
            - --service-account-issuer=https://kubernetes.default.svc.cluster.local
            - --service-account-key-file=/etc/kubernetes/pki/sa.pub
            - --service-account-signing-key-file=/etc/kubernetes/pki/sa.key
            - --service-cluster-ip-range={{ .ServiceCIDR }}
            - --tls-cert-file=/etc/kubernetes/pki/apiserver.crt
            - --tls-private-key-file=/etc/kubernetes/pki/apiserver.key              
          livenessProbe:
            failureThreshold: 3
            httpGet:
              path: /livez
              port: {{ .Port }}
              scheme: HTTPS
            periodSeconds: 10
            successThreshold: 1
            timeoutSeconds: 1
          readinessProbe:
            failureThreshold: 3
            httpGet:
              path: /readyz
              port: {{ .Port }}
              scheme: HTTPS
            periodSeconds: 10
            successThreshold: 1
            timeoutSeconds: 1
          resources:
            requests:
              cpu: 125m
              memory: 512Mi
          startupProbe:
            failureThreshold: 3
            httpGet:
              path: /livez
              port: {{ .Port }}
              scheme: HTTPS
            periodSeconds: 10
            successThreshold: 1
            timeoutSeconds: 1
          volumeMounts:
          - mountPath: /etc/kubernetes/pki
            name: kubernetes-pki
            readOnly: true
        - name: kube-scheduler
          image: registry.k8s.io/kube-scheduler:v{{ .Version }}
          command:
            - kube-scheduler
          args:
            - --authentication-kubeconfig=/etc/kubernetes/pki/scheduler.conf
            - --authorization-kubeconfig=/etc/kubernetes/pki/scheduler.conf
            - --bind-address=0.0.0.0
            - --kubeconfig=/etc/kubernetes/pki/scheduler.conf
            - --leader-elect=true
          livenessProbe:
            failureThreshold: 3
            httpGet:
              path: /healthz
              port: 10259
              scheme: HTTPS
            periodSeconds: 10
            successThreshold: 1
            timeoutSeconds: 1
          resources:
            requests:
              cpu: 125m
              memory: 256Mi
          startupProbe:
            failureThreshold: 3
            httpGet:
              path: /healthz
              port: 10259
              scheme: HTTPS
            periodSeconds: 10
            successThreshold: 1
            timeoutSeconds: 1
          volumeMounts:
            - mountPath: /etc/kubernetes/pki
              name: kubernetes-pki
              readOnly: true
        - name: kube-controller-manager
          image: registry.k8s.io/kube-controller-manager:v{{ .Version }}
          command:
            - kube-controller-manager
          args:
            - --allocate-node-cidrs=true
            - --authentication-kubeconfig=/etc/kubernetes/pki/controller-manager.conf
            - --authorization-kubeconfig=/etc/kubernetes/pki/controller-manager.conf
            - --bind-address=0.0.0.0
            - --client-ca-file=/etc/kubernetes/pki/ca.crt
            - --cluster-cidr={{ .ClusterCIDR }}
            - --cluster-name=dev
            - --cluster-signing-cert-file=/etc/kubernetes/pki/ca.crt
            - --cluster-signing-key-file=/etc/kubernetes/pki/ca.key
            - --controllers=*,bootstrapsigner,tokencleaner
            - --kubeconfig=/etc/kubernetes/pki/controller-manager.conf
            - --leader-elect=true
            - --requestheader-client-ca-file=/etc/kubernetes/pki/front-proxy-ca.crt
            - --root-ca-file=/etc/kubernetes/pki/ca.crt
            - --service-account-private-key-file=/etc/kubernetes/pki/sa.key
            - --service-cluster-ip-range={{ .ServiceCIDR }}
            - --use-service-account-credentials=true
          livenessProbe:
            failureThreshold: 3
            httpGet:
              path: /healthz
              port: 10257
              scheme: HTTPS
            periodSeconds: 10
            successThreshold: 1
            timeoutSeconds: 1
          resources:
            requests:
              cpu: 125m
              memory: 256Mi
          startupProbe:
            failureThreshold: 3
            httpGet:
              path: /healthz
              port: 10257
              scheme: HTTPS
            periodSeconds: 10
            successThreshold: 1
            timeoutSeconds: 1
          volumeMounts:
            - mountPath: /etc/kubernetes/pki
              name: kubernetes-pki
              readOnly: true
        - name: konnectivity-server
          image: registry.k8s.io/kas-network-proxy/proxy-server:v0.0.37
          command:
            - /proxy-server
          args:
            - --admin-port=8133
            - --agent-port=8132
            - --health-port=8134
            - --mode=grpc
            - --server-count=1
            - --server-port=0
            - --agent-namespace=tenant-dev
            - --agent-service-account=konnectivity-agent
            - --authentication-audience=system:konnectivity-server
            - --cluster-cert=/etc/kubernetes/pki/apiserver.crt
            - --cluster-key=/etc/kubernetes/pki/apiserver.key
            - --kubeconfig=/etc/kubernetes/pki/konnectivity-server.conf        
            - --uds-name=/run/konnectivity/konnectivity-server.socket
          volumeMounts:
            - mountPath: /etc/kubernetes/pki
              name: kubernetes-pki
              readOnly: true
            - mountPath: /run/konnectivity
              name: konnectivity-uds        
      volumes:
        - name: kubernetes-pki
          projected:
            sources:
              - secret:
                  name: ca
              - secret:
                  name: apiserver
              - secret:
                  name: apiserver-kubelet-client
              - secret:
                  name: front-proxy-ca
              - secret:
                  name: front-proxy-client
              - secret:
                  name: sa
              - secret:
                  name: kubeconfig-scheduler
              - secret:
                  name: kubeconfig-controller
              - secret:
                  name: kubeconfig-konnectivity
        - name: konnectivity-uds
          emptyDir:
            medium: Memory
`
