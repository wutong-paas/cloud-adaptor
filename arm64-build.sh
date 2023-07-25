#! /bin/bash

export NAMESPACE=wutong-adaptor
export VERSION=v1.4.0-arm64

CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o ./bin/arm64/cloud-adaptor ./cmd/cloud-adaptor

docker build --build-arg TARGET=arm64 -t swr.cn-southwest-2.myhuaweicloud.com/wutong/${NAMESPACE}:${VERSION} -f Dockerfile.local . 
docker push swr.cn-southwest-2.myhuaweicloud.com/wutong/${NAMESPACE}:${VERSION}

rm -rf ./bin/*