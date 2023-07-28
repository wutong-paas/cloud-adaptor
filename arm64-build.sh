#! /bin/bash

export IMAGE_REGISTRY=swr.cn-southwest-2.myhuaweicloud.com/wutong
export IMAGE_NAME=wutong-adaptor
export VERSION=v1.4.0-arm64

CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -a -o ./bin/arm64/cloud-adaptor ./cmd/cloud-adaptor

docker build -t ${IMAGE_REGISTRY}/${IMAGE_NAME}:${VERSION} -f Dockerfile.local . 
docker push ${IMAGE_REGISTRY}/${IMAGE_NAME}:${VERSION}

rm -rf ./bin/*