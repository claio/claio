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

import "fmt"

func (s *CertificateFactory) Check() error {
	log := s.Factory.Log
	log.Info("   check secrets ...")
	// ca
	_, caChanged, err := s.GetCa(false)
	if err != nil {
		return fmt.Errorf("failed to get ca")
	}
	// apiserver (force renew if CA has changed)
	_, _, err = s.GetApiserver(caChanged)
	if err != nil {
		return fmt.Errorf("failed to get apiserver")
	}
	// apiserver-kubelet-client (force renew if CA has changed)
	_, _, err = s.GetApiserverKubeletClient(caChanged)
	if err != nil {
		return fmt.Errorf("failed to get apiserver-kubelet-client")
	}
	// front-proxy-ca
	_, frontProxyCaChanged, err := s.GetFrontProxyCa(false)
	if err != nil {
		return fmt.Errorf("failed to get front-proxy-ca")
	}
	// front-proxy-client
	_, _, err = s.GetFrontProxyClient(frontProxyCaChanged)
	if err != nil {
		return fmt.Errorf("failed to get front-proxy-client")
	}
	// sa
	_, _, err = s.GetSa(false)
	if err != nil {
		return fmt.Errorf("failed to get sa")
	}
	// kubeconfig-admin
	/*
		_, _, err = s.GetKubeconfigAdmin(false)
		if err != nil {
			return fmt.Errorf("failed to get kubeconfig-admin")
		}
	*/

	return nil
}
