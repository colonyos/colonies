#!/bin/bash

namespace="colonies"

echo "-> Creating namespace" 
kubectl apply -f namespace.yaml 

echo "-> Creating TLS secret" 
mkdir tls-cert
chmod 0700 "./tls-cert"
cd "./tls-cert"

openssl req -nodes -new -x509 -days 10000 -keyout ca.key -out ca.crt -subj "/CN=ColonyOS Colonies CA"
openssl genrsa -out colonies-cert-tls.key 2048
openssl req -new -key colonies-cert-tls.key -subj "/CN=colonies-deployment.${namespace}.svc" \
    | openssl x509 -days 10000 -req -CA ca.crt -CAkey ca.key -CAcreateserial -out colonies-cert-tls.crt

cd ..

echo "-> Creating Kubernetes TLS keys and secret"
kubectl -n ${namespace} create secret tls colonies-cert-tls --cert ./tls-cert/colonies-cert-tls.crt --key ./tls-cert/colonies-cert-tls.key
