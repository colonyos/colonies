#!/bin/bash

namespace="rocksigma-computer-colonies"
echo "upgrading ${namespace} ..."
helm upgrade compute -f values.yaml -n ${namespace} --wait .
