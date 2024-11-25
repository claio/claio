#!/bin/bash

API_PORT=9443
REGISTRY_PORT=5005
remote_docker=true

function usage() {
    echo "$0 up|down"
    exit 1
}

if [ -z "$1" ]; then
    usage
fi

# check if we have a local docker daemon running
if pidof dockerd > /dev/null; then
    remote_docker=false
fi

# define ctlptl manifests
read -r -d '' manifests <<EOF
apiVersion: ctlptl.dev/v1alpha1
kind: Registry
name: claio-registry
port: ${REGISTRY_PORT}
---
apiVersion: ctlptl.dev/v1alpha1
kind: Cluster
product: kind
name: kind-claio
kindV1Alpha4Cluster:
    networking:
        apiServerPort: ${API_PORT}
        podSubnet: 192.168.0.0/17
        serviceSubnet: 192.168.128.0/17
    nodes:
        - role: control-plane
        - role: worker
EOF

function portforward() {
    echo "create portforward for $1 ($2)"
    socat "TCP-LISTEN:$2,reuseaddr,fork" \
        EXEC:"'docker exec -i claio-portforward socat STDIO TCP4:localhost:$2'" 2>/dev/null 1>/dev/null &
}

case "$1" in
    up)        
        if [ $remote_docker == "true" ]; then
            echo -n "start portforward service: "        
            docker run -d -it --name claio-portforward --net=host --entrypoint=/bin/sh \
                alpine/socat -c "while true; do sleep 1000; done"
            portforward "api-server" ${API_PORT} 
            portforward "registry" ${REGISTRY_PORT}
            sleep 3
        fi

        echo "create registry and kind cluster"
        echo "$manifests" | ctlptl apply -f -      
        ;;
    down)
        #tilt down
        echo "$manifests" | ctlptl delete -f -        
        pkill socat 2>/dev/null 1>/dev/null
        docker rm -f claio-portforward 2>/dev/null 1>/dev/null
        ;;
    *)
        usage
        ;;
esac

