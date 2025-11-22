#!/bin/bash

set -euo pipefail

shopt -s extglob

rm -rf !("broker-san.cnf"|"generate-certificates.sh")

read -sp "Password: " PASSWORD
export KEYSTORE=kafka.server.keystore.jks
export TRUSTSTORE=kafka.server.truststore.jks

# Private key of CA
openssl genrsa -out ca.key 4096

# Self-signed CA certificate
openssl req -x509 -new -nodes \
  -key ca.key \
  -sha256 -days 3650 \
  -out ca.crt \
  -subj "/CN=ktea-CA"

# Generate private key for broker
openssl genrsa -out server.key 2048

# Generate CSR with SAN from broker-san.cnf
openssl req -new \
  -key server.key \
  -out kafka-broker.csr \
  -config broker-san.cnf

# Sign the broker certificate with SAN
openssl x509 -req \
  -in kafka-broker.csr \
  -CA ca.crt \
  -CAkey ca.key \
  -CAcreateserial \
  -out kafka-broker-signed.crt \
  -days 365 \
  -sha256 \
  -extensions req_ext \
  -extfile broker-san.cnf


# Create keystore and import certs

# Create keystore and insert private key
openssl pkcs12 -export \
  -in kafka-broker-signed.crt \
  -inkey server.key \
  -name kafka-broker \
  -out kafka-broker.p12 \
  -password pass:$PASSWORD

keytool -importkeystore \
  -deststorepass $PASSWORD \
  -destkeystore $KEYSTORE \
  -srckeystore kafka-broker.p12 \
  -srcstoretype PKCS12 \
  -srcstorepass $PASSWORD \
  -alias kafka-broker \
  -noprompt

# Import CA
keytool -import \
  -alias CARoot \
  -file ca.crt \
  -keystore $KEYSTORE \
  -storepass $PASSWORD \
  -noprompt

echo $PASSWORD >keystore_creds
echo $PASSWORD >sslkey_creds
echo $PASSWORD >truststore_creds

cat >server-jaas.conf <<EOF
KafkaServer {
  org.apache.kafka.common.security.plain.PlainLoginModule required
  username="test"
  password="testtest"
  user_test="testtest"
  user_bob="bobbob"
  user_alice="alicealice";

  org.apache.kafka.common.security.scram.ScramLoginModule required
  username="admin"
  password="adminadmin";
};
KafkaClient {
  org.apache.kafka.common.security.plain.PlainLoginModule required
  username="test"
  password="testtest"
  user_test="testtest"
  user_bob="bobbob"
  user_alice="alicealice";
};
Server {
  org.apache.kafka.common.security.plain.PlainLoginModule required
  username="test"
  password="testtest"
  user_test="testtest";

  org.apache.kafka.common.security.scram.ScramLoginModule required
  username="admin"
  password="adminadmin";
};
Client {
  org.apache.kafka.common.security.plain.PlainLoginModule required
  username="test"
  password="testtest"
  user_test="testtest";
};
EOF
