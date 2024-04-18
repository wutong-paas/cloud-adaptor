#! /bin/bash

WUTONG_REGISTRY=${WUTONG_REGISTRY:-'swr.cn-southwest-2.myhuaweicloud.com/wutong'}
IMAGE_NAME=wutong-adaptor
VERSION=v1.12.0

go mod download

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o bin/amd64/cloud-adaptor ./cmd/cloud-adaptor
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 GO111MODULE=on go build -a -o bin/arm64/cloud-adaptor ./cmd/cloud-adaptor

docker buildx use swrbuilder || docker buildx create --use --name swrbuilder --driver docker-container --driver-opt image=swr.cn-southwest-2.myhuaweicloud.com/wutong/buildkit:stable
docker buildx build --platform linux/amd64,linux/arm64 --push -t ${WUTONG_REGISTRY}/${IMAGE_NAME}:${VERSION} -f Dockerfile.local . 
# docker buildx rm swrbuilder

rm -rf ./bin/*
