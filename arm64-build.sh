#! /bin/bash

WUTONG_REGISTRY=${WUTONG_REGISTRY:-'swr.cn-southwest-2.myhuaweicloud.com/wutong'}
IMAGE_NAME=wutong-adaptor
VERSION=v1.13.0-arm64

CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -a -o ./bin/arm64/cloud-adaptor ./cmd/cloud-adaptor

docker build -t ${WUTONG_REGISTRY}/${IMAGE_NAME}:${VERSION} -f Dockerfile.local . 
docker push ${WUTONG_REGISTRY}/${IMAGE_NAME}:${VERSION}

rm -rf ./bin/*