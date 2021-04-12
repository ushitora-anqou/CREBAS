#!/bin/bash

PROJECT_DIR=$(cd $(dirname $0); pwd)
echo $PROJECT_DIR

cd cmd/appdaemon
go build
cp appdaemon /tmp/.
cd $PROJECT_DIR

go test -v ./...
