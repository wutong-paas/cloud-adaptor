#! /bin/bash

export NAMESPACE=wutong-adaptor
export VERSION=v1.4.0-arm64

CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o ./bin/cloud-adaptor ./cmd/cloud-adaptor
docker build . -t swr.cn-southwest-2.myhuaweicloud.com/wutong/${NAMESPACE}:${VERSION} -f Dockerfile.arm64
docker push swr.cn-southwest-2.myhuaweicloud.com/wutong/${NAMESPACE}:${VERSION}