# claio
A hosted control plane and machine controller (all in one).

## Development

```sh
# start
./scripts/dev-cluster.sh start
./tilt up

# stop
./tilt down --delete-namespaces
./scripts/dev-cluster.sh stop
```

### Installation History (not needed anymore)

```sh
go mod init claio
kubebuilder init --domain github.com
kubebuilder create api --resource --controller --group claio --version v1alpha1 --kind ControlPlane
kubebuilder create api --resource --controller --group claio --version v1alpha1 --kind Machine
```