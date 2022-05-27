#!/bin/bash

namespace="rocksigma-computer-colonies"

kubectl create namespace ${namespace}
kubectl -n ${namespace} create secret docker-registry prvdockerreg --docker-server=https://registry.rocksigma.computer --docker-username=rsdev --docker-password=5RC5wefddeuYw6ij2x9eVGa5p3XT

helm install ${namespace} -f values.yaml -n ${namespace} .
