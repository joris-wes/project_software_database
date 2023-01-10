#!/bin/sh

set -xe

# Rebuild the image

docker-compose --env-file .env build $1
docker-compose --env-file .env create $1
docker-compose --env-file .env start $1