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
	"time"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/apps/v1"
)

func (c *ControlPlane) CreateClaioDeployment() error {
	deploymentYaml, err := c.ToYaml(controlplaneTemplate, c.Object.Spec)
	if err != nil {
		return fmt.Errorf("error generating yaml: %s", err)
	}
	return c.CreateDeployment("claio", deploymentYaml)
}

func (c *ControlPlane) UpdateClaioDeployment() error {
	deploymentYaml, err := c.ToYaml(controlplaneTemplate, c.Object.Spec)
	if err != nil {
		return fmt.Errorf("error generating yaml: %s", err)
	}
	return c.UpdateDeployment("claio", deploymentYaml)
}

func (c *ControlPlane) DeleteClaioDeployment() error {
	deploymentYaml, err := c.ToYaml(controlplaneTemplate, c.Object.Spec)
	if err != nil {
		return fmt.Errorf("error generating yaml: %s", err)
	}
	return c.DeleteDeployment("claio", deploymentYaml)
}

func (c *ControlPlane) GetClaioDeployment() (*v1.Deployment, error) {
	deployment, err := c.GetDeployment("claio")
	if err != nil {
		return nil, fmt.Errorf("error getting deployment: %s", err)
	}
	return deployment, nil
}

func (c *ControlPlane) ReconcileDeployment(apiDirty bool, mode string) error {
	c.LogHeader("check deployment ...")
	deployment, err := c.GetClaioDeployment()
	if err != nil {
		c.LogError(err, "failed to retreive claio deployment")
		return err
	}

	// stop deployment, controlplane wants to stop
	if mode == c.STATUS_WANTDOWN {
		if deployment == nil {
			c.LogInfo("deployment already deleted")
		} else {
			if err := c.stopPods(deployment); err != nil {
				return err
			}
		}
		return nil
	}

	if deployment == nil {
		c.LogInfo("create claio deployment")
		if err := c.CreateClaioDeployment(); err != nil {
			c.LogError(err, "failed to create deployment")
			return err
		}
		return nil
	}

	if apiDirty || !c.isEqual() {
		c.LogInfo("structural changes detected - need to stop control-plane")
		if err := c.stopDeployment(); err != nil {
			c.LogError(err, "failed to stop deployment")
			return err
		}

		// the deployment will be startet with the next reconcilation run
		return nil
	}
	return nil
}

func (c *ControlPlane) isEqual() bool {
	return reflect.DeepEqual(c.Object.Spec, c.Object.Status.TargetSpec)
}

func (c *ControlPlane) stopDeployment() error {
	c.LogInfo("stop deployment")
	if err := c.DeleteClaioDeployment(); err != nil {
		c.LogError(err, "failed to delete deployment")
		return err
	}
	loop := 0
	for {
		if loop > 10 {
			return fmt.Errorf("failed to delete deployment (loop)")
		}
		depl, err := c.GetClaioDeployment()
		if err == nil && depl == nil {
			break
		}
		time.Sleep(3 * time.Second)
		loop++
	}
	c.LogInfo("deployment stopped")
	return nil
}

func (c *ControlPlane) stopPods(deployment *appsv1.Deployment) error {
	c.LogInfo("scale replicas down to 0")
	*deployment.Spec.Replicas = 0
	if err := c.UpdateClaioDeployment(); err != nil {
		return fmt.Errorf("failed to set deployments replicas to 0")
	}
	return nil
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
            - --etcd-prefix=/tenant-{{ .Name }}
            - --etcd-servers=http://localhost:2379
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
        - name: kine
          image: rancher/kine:v0.13.2
          args:
            - --endpoint 
            - "nats://nats.claio-system.svc?noEmbed&bucket=tenant-sample"
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
