# claio

A hosted control plane and machine controller manager (all in one).

> [!IMPORTANT]
> This project is in a very early stage. At the moment it is not even ready for development.

## Development

```sh
# start
./scripts/cluster.sh up
./tilt up

# stop
./tilt down --delete-namespaces
./scripts/cluster.sh down
```

### Installation History (not needed anymore)

```sh
go mod init claio
kubebuilder init --domain github.com
kubebuilder create api --resource --controller --group claio --version v1alpha1 --kind ControlPlane
kubebuilder create api --resource --controller --group claio --version v1alpha1 --kind Machine
```
