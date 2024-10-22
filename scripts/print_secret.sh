#!/bin/sh

set -e

function usage() {
    echo "Usage: $0 <secret>"
    exit 1
}

secret=$1
key=$2

if [ -z "$secret" ]; then
    usage
fi

if [[ -z "$key" ||  $key = "crt" ]]; then
    key="crt"
    kubectl get secret $secret -o jsonpath='{.data.'${secret}'\.'${key}'}' | 
        base64 -d | 
        openssl x509 -text -nocert | 
        grep -v "Signature Value:" |
        egrep -v '^[[:space:]]*[0-9a-z][0-9a-z]\:' 
else
    kubectl get secret $secret -o jsonpath='{.data.'${key}'}' | base64 -d 
fi