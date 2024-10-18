#!/bin/sh

set -e

function usage() {
    echo "Usage: $0 <secret>"
    exit 1
}

secret=$1

if [ -z "$secret" ]; then
    usage
fi

kubectl get secret $secret -o jsonpath='{.data.'${secret}'\.crt}' | 
    base64 -d | 
    openssl x509 -text -nocert | 
    grep -v "Signature Value:" |
    egrep -v '^[[:space:]]*[0-9a-z][0-9a-z]\:' 