#!/bin/bash
set -e

source /configuration/colonyos.env
source /configuration/minio.env

echo "source /configuration/colonyos.env" >> /home/vscode/.bashrc
echo "source /configuration/minio.env" >> /home/vscode/.bashrc

echo "loaded environment variables"

go install github.com/minio/mc@latest

until (mc alias set myminio http://minio:9000 ${MINIO_ROOT_USER} ${MINIO_ROOT_PASSWORD}); do echo 'Waiting for MinIO...'; sleep 3; done;
echo 'MinIO is ready. Setting up...';
mc admin user add myminio ${AWS_S3_ACCESSKEY} ${AWS_S3_SECRETKEY};
mc admin policy attach myminio readwrite --user=${AWS_S3_ACCESSKEY};
mc mb myminio/${AWS_S3_BUCKET};
echo 'MinIO setup completed.';