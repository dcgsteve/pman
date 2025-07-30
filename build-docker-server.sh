#!/bin/bash

IMG=dcgsteve/pman-server
VERSION=$(grep 'Version = ' cli/main.go | sed 's/.*Version = "\(.*\)".*/\1/')

docker build . -t ${IMG}:${VERSION}
docker tag ${IMG}:${VERSION} ${IMG}:latest

# Only push if /p flag is passed
if [ "${1,,}" = "/p" ]; then
    echo "Pushing to Docker Hub..."
    docker push ${IMG}:${VERSION}
    docker push ${IMG}:latest
else
    echo "Use './$0 /p' to also push to Docker Hub"
fi

