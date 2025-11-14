#!/bin/bash
set -e

source /configuration/postgres.env
exec docker-entrypoint.sh postgres
