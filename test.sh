#!/bin/bash

PROJECT_DIR=$(cd $(dirname $0); pwd)
echo $PROJECT_DIR

sudo docker build -t crebas-test test/docker/
if [ $? -ne 0 ]; then
    echo "Failed to build test container"
    exit 1
fi

sudo docker run -v $PROJECT_DIR:/CREBAS --rm --cap-add=NET_ADMIN crebas-test