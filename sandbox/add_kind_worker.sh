#!/bin/bash

set -e

VERSION="v1.27.1"
NODE_NAME="node01"

kubectl config use-context kubernetes-admin@tenant-sample

# cleanup
docker kill $NODE_NAME || true
docker rm $NODE_NAME || true

# 1. start docker container running kindest/node
docker run \
    --restart on-failure \
    -v /boot:/boot:ro \
    -v /dev/mapper:/dev/mapper \
    -v /lib/modules:/lib/modules:ro \
    --privileged \
    -h $NODE_NAME \
    --add-host "$NODE_NAME:127.0.0.1" \
    -d \
    --tmpfs /run \
    --tmpfs /tmp \
    --net=host \
    --security-opt seccomp=unconfined \
    --security-opt apparmor=unconfined \
    --security-opt label=disable \
    -v /var \
    --name $NODE_NAME \
    --label io.x-k8s.kind.cluster=claio \
    --label io.x-k8s.kind.role=worker \
    --env KIND_EXPERIMENTAL_CONTAINERD_SNAPSHOTTER \
    kindest/node:$VERSION

sleep 3

docker cp /usr/bin/socat $NODE_NAME:/usr/bin/socat

docker exec -i $NODE_NAME bash -c "systemctl status containerd"
docker exec $NODE_NAME bash -c "mkdir -p /root/.kube"
kubectl config view --raw=true | docker exec -i $NODE_NAME bash -c "cat - > /root/.kube/config" 

cmd=$(docker exec $NODE_NAME bash -c "kubeadm token create --print-join-command")
echo $cmd
#docker exec $NODE_NAME bash -c "$cmd"
