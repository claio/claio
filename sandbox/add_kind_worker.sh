#!/bin/bash

set -e

VERSION="v1.27.1"
NODE_NAME="kind-worker-sample"

kubectl config use-context kubernetes-admin@tenant-sample

# cleanup
docker kill $NODE_NAME || true
docker rm $NODE_NAME || true

# 1. start docker container running kindest/node
docker run \
    --restart on-failure \
    -v /lib/modules:/lib/modules:ro \
    --privileged \
    -h $NODE_NAME \
    -d \
    --network kind \
    --network-alias $NODE_NAME \
    --tmpfs /run \
    --tmpfs /tmp \
    --security-opt seccomp=unconfined \
    --security-opt apparmor=unconfined \
    --security-opt label=disable \
    -v /var \
    --name $NODE_NAME \
    --label io.x-k8s.kind.cluster=kind \
    --label io.x-k8s.kind.role=worker \
    --env KIND_EXPERIMENTAL_CONTAINERD_SNAPSHOTTER \
    kindest/node:$VERSION

JOIN=$(kubeadm token create --print-join-command)

echo $JOIN

#docker exec $NODE_NAME bash -c "$JOIN"