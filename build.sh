#! /bin/bash

export NAMESPACE=wutong-adaptor
export VERSION=v1.4.0
docker buildx create --use --name adaptorbuilder || docker buildx use adaptorbuilder
docker buildx build --platform linux/amd64,linux/arm64 --push -t swr.cn-southwest-2.myhuaweicloud.com/wutong/${NAMESPACE}:${VERSION} -f Dockerfile.multiarch . 
# docker buildx rm adaptorbuilder