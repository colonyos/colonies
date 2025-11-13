#!/bin/sh

if [ -f /configuration/postgres.env ]; then
    exit 0
fi

DB_PASSWORD=$(head -c 16 /dev/urandom | base64)

echo "export POSTGRES_PASSWORD=$DB_PASSWORD" >> /configuration/postgres.env
echo "export COLONIES_DB_PASSWORD=$DB_PASSWORD" >> /configuration/colonyos.env

COLONIES_PRVKEY=$(colonies security generate 2>&1 | grep 'PrvKey=' | awk -F'PrvKey=' '{print $2}')
echo "export COLONIES_PRVKEY=$COLONIES_PRVKEY" >> /configuration/colonyos.env

SERVER_KEYPAIR=$(colonies security generate 2>&1)
SERVER_ID=$(echo "$SERVER_KEYPAIR" | grep -o 'Id=[^ ]*' | awk -F'=' '{print $2}')
SERVER_PRVKEY=$(echo "$SERVER_KEYPAIR" | grep -o 'PrvKey=[^ ]*' | awk -F'=' '{print $2}')

echo "export COLONIES_SERVER_ID=$SERVER_ID" >> /configuration/colonyos.env
echo "export COLONIES_SERVER_PRVKEY=$SERVER_PRVKEY" >> /configuration/colonyos.env

COLONY_KEYPAIR=$(colonies security generate 2>&1)
COLONY_ID=$(echo "$COLONY_KEYPAIR" | grep -o 'Id=[^ ]*' | awk -F'=' '{print $2}')
COLONY_PRVKEY=$(echo "$COLONY_KEYPAIR" | grep -o 'PrvKey=[^ ]*' | awk -F'=' '{print $2}')

echo "export COLONIES_COLONY_ID=$COLONY_ID" >> /configuration/colonyos.env
echo "export COLONIES_COLONY_PRVKEY=$COLONY_PRVKEY" >> /configuration/colonyos.env

MINIO_ROOT_PASSWORD=$(head -c 16 /dev/urandom | base64)
echo "export MINIO_ROOT_PASSWORD=$MINIO_ROOT_PASSWORD" >> /configuration/minio.env

AWS_S3_ACCESSKEY=$(head -c 16 /dev/urandom | base64 | tr -d '=.' )
AWS_S3_SECRETKEY=$(head -c 32 /dev/urandom | base64 | tr -d '=.' )

echo "export AWS_S3_ACCESSKEY=$AWS_S3_ACCESSKEY" >> /configuration/minio.env
echo "export AWS_S3_SECRETKEY=$AWS_S3_SECRETKEY" >> /configuration/minio.env
