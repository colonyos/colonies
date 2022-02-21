#!/bin/bash

export SERVERID="9289dfccedf27392810b96968535530bb69f90afe7c35738e0e627f3810d943e"
export SERVERPORT="50051"
export TLSCERT="/cert/cert.pem"
export TLSKEY="/cert/key.pem"
export DBHOST="10.0.0.240"
export DBPORT="5432"
export DBUSER="postgres"
export DBPASSWORD="rFcLGNkgsNtksg6Pgtn9CumL4xXBQ7"
export TZ=Europe/Stockholm
export VERBOSE="true"

docker run -d --restart unless-stopped -v $PWD/../../cert:/cert -e SERVERID=$SERVERID -e SERVERPORT=$SERVERPORT -e TLSCERT=$TLSCERT -e TLSKEY=$TLSKEY -e DBHOST=$DBHOST -e DBPORT=$DBPORT -e DBUSER=$DBUSER -e DBPASSWORD=$DBPASSWORD -e TZ=$TZ -e VERBOSE=$VERBOSE colonyos/colonies:latest
