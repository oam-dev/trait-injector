#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

export APP="${1:-injector-trait}"
export NAMESPACE="${2:-default}"

output_dir="_generated/"

mkdir -p ${output_dir}/

# CREATE THE PRIVATE KEY FOR OUR CUSTOM CA
openssl genrsa -out ${output_dir}/ca.key 2048

cat > ${output_dir}/ca_config.txt <<EOF
[ req ]
default_bits       = 2048
default_md         = sha512
prompt             = no
encrypt_key        = yes

# base request
distinguished_name = req_distinguished_name

# extensions
req_extensions     = v3_req

# distinguished_name
[ req_distinguished_name ]
countryName            = "CN"                     # C=
stateOrProvinceName    = "ZJ"                     # ST=
localityName           = "HZ"                     # L=
postalCode             = "000000"                 # L/postalcode=
streetAddress          = "oam"                    # L/street=
organizationName       = "OAM"                    # O=
organizationalUnitName = "oam"                    # OU=
commonName             = "oam.dev"                # CN=
emailAddress           = "no-reply@oam.dev"       # CN/emailAddress=

# req_extensions
[ v3_req ]
# The subject alternative name extension allows various literal values to be
# included in the configuration file
# http://www.openssl.org/docs/apps/x509v3_config.html
subjectAltName  = DNS:www.oam.dev,DNS:oam.dev # multidomain certificate
EOF

# GENERATE A CA CERT WITH THE PRIVATE KEY
# This is the CA to sign the webhook's serving cert.
openssl req -new -x509 -key ${output_dir}/ca.key -out ${output_dir}/ca.crt -config ${output_dir}/ca_config.txt

# CREATE THE PRIVATE KEY FOR THE WEBHOOK
openssl genrsa -out ${output_dir}/service-injector.key 2048

cat > ${output_dir}/csr_config.txt <<EOF
[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name
[ req_distinguished_name ]
[ v3_req ]
basicConstraints=CA:FALSE
subjectAltName=@alt_names
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth

[ alt_names ]
DNS.1 = ${APP}
DNS.2 = ${APP}.${NAMESPACE}
DNS.3 = ${APP}.${NAMESPACE}.svc
DNS.4 = ${APP}.${NAMESPACE}.svc.cluster.local
EOF
# CREATE A CSR FROM THE CONFIGURATION FILE AND OUR PRIVATE KEY
openssl req -new -key ${output_dir}/service-injector.key -subj "/CN=${APP}.${NAMESPACE}.svc" -out ${output_dir}/admission.csr -config ${output_dir}/csr_config.txt

# CREATE THE CERT SIGNED BY THE CSR AND THE CA
openssl x509 -req -days 365 -in ${output_dir}/admission.csr -CA ${output_dir}/ca.crt -CAkey ${output_dir}/ca.key -CAcreateserial -out ${output_dir}/service-injector.pem

# Create certificates for Webhook to consume as a secret
# kubectl create secret generic ${APP} -n ${NAMESPACE} \
#   --from-file=${output_dir}/service-injector.key \
#   --from-file=${output_dir}/service-injector.pem

# Set 'caBundle' in Helm chart
# sed -i .bak 's/_CABundle_/'"$(cat ${output_dir}/ca.crt | base64 | tr -d '\n'}")'/g'
