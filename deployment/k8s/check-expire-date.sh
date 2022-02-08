#!/bin/bash
cat ./tls-cert/colonies-cert-tls.crt | openssl x509 -noout -enddate
