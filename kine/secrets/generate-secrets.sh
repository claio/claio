#!/bin/sh

set -e

cd secrets
mkdir -p /tmp/claio

echo "Generate self signed root CA cert"
openssl req -new -x509 -days 3650 -config server_openssl.cnf -keyout /tmp/claio/ca.key -out /tmp/claio/ca.crt -nodes

echo "Create a private key for the server"
openssl genrsa -out /tmp/claio/server.key 2048

echo "Generate server CSR with SAN"
openssl req -new -key /tmp/claio/server.key -out /tmp/claio/server.csr -config server_openssl.cnf

echo "Sign the server CSR with CA"
openssl x509 -req -in /tmp/claio/server.csr -CA /tmp/claio/ca.crt -CAkey /tmp/claio/ca.key \
        -CAcreateserial -out /tmp/claio/server.crt -days 365 -extfile server_openssl.cnf -extensions v3_req

echo "Verify if it's a SAN cert"
openssl x509 -in /tmp/claio/server.crt -text -noout | grep -A1 "Subject Alternative Name"

cp mariadb-ssl.cnf /tmp/claio/my.cnf
if [ ! -f mariadb-root-password ]; then
        tr -dc 'A-Za-z0-9!?%=' < /dev/urandom | head -c 24 > mariadb-root-password
fi
cp mariadb-root-password /tmp/claio/root-password

cat <<EOF > /tmp/claio/kustomization.yaml
namespace: claio-system
namePrefix: claio-kine-

secretGenerator:
  - name: mariadb
    options:
      disableNameSuffixHash: true
    files:
      - my.cnf
      - root-password
      - ca.crt
      - server.crt
      - server.key
EOF

kustomize build /tmp/claio > ../secrets.yaml
rm -rf /tmp/claio
