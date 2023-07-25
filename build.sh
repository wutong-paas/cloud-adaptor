#! /bin/bash

export NAMESPACE=wutong-adaptor
export VERSION=v1.4.0

go mod download

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o bin/amd64/cloud-adaptor ./cmd/cloud-adaptor
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 GO111MODULE=on go build -a -o bin/arm64/cloud-adaptor ./cmd/cloud-adaptor

docker buildx use gobuilder || docker buildx create --use --name gobuilder
docker buildx build --platform linux/amd64,linux/arm64 --push -t swr.cn-southwest-2.myhuaweicloud.com/wutong/${NAMESPACE}:${VERSION} -f Dockerfile.local . 
# docker buildx rm gobuilder

rm -rf ./bin/*