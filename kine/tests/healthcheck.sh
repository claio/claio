#!/bin/bash

echo "Check etcd/kine health ..."
cmd="etcdctl get /registry/health"
echo "$cmd --> expect health=true"
status=$($cmd)
echo "--> " $status
echo "$status" | tail -1 | grep -q "true"
