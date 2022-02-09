#!/bin/bash

namespace="colonies"

echo "Installing Colonies on Kubernetes ..."
echo "-> Creating Colonies namespace" 
kubectl create namespace ${namespace}

echo ""
echo "-> Creating TLS secret" 
mkdir tls-cert
chmod 0700 "./tls-cert"
cd "./tls-cert"

openssl req -nodes -new -x509 -days 10000 -keyout ca.key -out ca.crt -subj "/CN=ColonyOS Colonies CA"
openssl genrsa -out colonies-cert-tls.key 2048
openssl req -new -key colonies-cert-tls.key -subj "/CN=colonies-deployment.${namespace}.svc" \
    | openssl x509 -days 10000 -req -CA ca.crt -CAkey ca.key -CAcreateserial -out colonies-cert-tls.crt

cd ..
echo ""
echo "-> Creating Kubernetes TLS keys and secret"
kubectl -n ${namespace} create secret tls colonies-cert-tls --cert ./tls-cert/colonies-cert-tls.crt --key ./tls-cert/colonies-cert-tls.key

echo ""
echo "-> Deploying TimescaleDB"
kubectl -n ${namespace} apply -f postgres-configmap.yaml
kubectl -n ${namespace} apply -f postgres-storage.yaml
kubectl -n ${namespace} apply -f postgres-deployment.yaml
kubectl -n ${namespace} apply -f postgres-service.yaml

echo ""
echo "-> Deploying Colonies server"
kubectl -n ${namespace} apply -f colonies-configmap.yaml
kubectl -n ${namespace} apply -f colonies-deployment.yaml
