#! /bin/bash

export NAMESPACE=wutong-adaptor
export VERSION=v1.4.0-amd64

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/amd64/cloud-adaptor ./cmd/cloud-adaptor

docker build -t swr.cn-southwest-2.myhuaweicloud.com/wutong/${NAMESPACE}:${VERSION} -f Dockerfile.local . 
docker push swr.cn-southwest-2.myhuaweicloud.com/wutong/${NAMESPACE}:${VERSION}

rm -rf ./bin/*