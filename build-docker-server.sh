#!/bin/bash

IMG=dcgsteve/pman-server
VER=1.0.1

docker build . -t ${IMG}:${VER}
docker tag ${IMG}:${VER} ${IMG}:latest

# Only push if /p flag is passed
if [ "${1,,}" = "/p" ]; then
    echo "Pushing to Docker Hub..."
    docker push ${IMG}:${VER}
    docker push ${IMG}:latest
else
    echo "Use './$0 /p' to also push to Docker Hub"
fi

