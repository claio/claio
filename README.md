# claio
A hosted control plane and machine controller (all in one).

## Installation History (not needed anymore)

```sh
go mod init claio
kubebuilder init --domain github.com
kubebuilder create api --resource --controller --group claio --version v1alpha1 --kind ControlPlane
kubebuilder create api --resource --controller --group claio --version v1alpha1 --kind Machine
```