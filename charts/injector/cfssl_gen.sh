#!/bin/bash

# References:
#   https://coreos.com/os/docs/latest/generate-self-signed-certificates.html
#   https://github.com/coreos/etcd-operator/tree/master/example/tls/certs
# command to check certificate:
#   openssl x509 -in certificate.pem -text -noout

set -o errexit
set -o nounset
set -o pipefail

export APP="${1:-trait-injector}"
export NAMESPACE="${2:-default}"

mkdir -p _generated/
pushd _generated/

cat >ca-config.json <<EOF
{
    "signing": {
        "default": {
            "expiry": "43800h"
        },
        "profiles": {
            "server": {
                "expiry": "43800h",
                "usages": [
                    "signing",
                    "key encipherment",
                    "server auth"
                ]
            },
            "client": {
                "expiry": "43800h",
                "usages": [
                    "signing",
                    "key encipherment",
                    "client auth"
                ]
            }
        }
    }
}
EOF

cat >ca-csr.json <<EOF
{
    "CN": "My own CA",
    "key": {
        "algo": "rsa",
        "size": 2048
    },
    "names": [
        {
            "C": "US",
            "L": "CA",
            "O": "My Company Name",
            "ST": "San Francisco",
            "OU": "Org Unit 1",
            "OU": "Org Unit 2"
        }
    ]
}
EOF

cat >server.json <<EOF
{
    "CN": "server",
    "hosts": [
        "${APP}",
        "${APP}.${NAMESPACE}",
        "${APP}.${NAMESPACE}.svc",
        "${APP}.${NAMESPACE}.svc.cluster.local",
        "localhost",
        "127.0.0.1"
    ],
    "key": {
        "algo": "rsa",
        "size": 2048
    },
    "names": [
        {
            "C": "US",
            "L": "CA",
            "ST": "San Francisco"
        }
    ]
}
EOF


# It generates:
#   ca-key.pem
#   ca.csr
#   ca.pem
cfssl gencert -initca ca-csr.json | cfssljson -bare ca -

# It generates:
#   server-key.pem
#   server.csr
#   server.pem
cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json -profile=server server.json | cfssljson -bare server

popd

# Fill placeholders in Chart
sourcedir=$(dirname "$0")
sed -i.bak \
    -e 's/_CABundle_/'"$(cat ca.pem | base64 | tr -d '\n'})"'/g' \
    -e 's/_injectorKey_/'"$(cat server-key.pem | base64 | tr -d '\n'})"'/g' \
    -e 's/_injectorCrt_/'"$(cat server.pem | base64 | tr -d '\n'})"'/g' \
    ${sourcedir}/values.yaml

