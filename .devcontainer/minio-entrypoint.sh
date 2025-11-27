#!/bin/sh

source /configuration/minio.env
minio server /var/lib/minio/data --console-address :9001