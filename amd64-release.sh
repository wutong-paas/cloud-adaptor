#! /bin/bash

export NAMESPACE=wutong-adaptor
export VERSION=v1.0.0-stable

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/cloud-adaptor ./cmd/cloud-adaptor
docker build . -t swr.cn-southwest-2.myhuaweicloud.com/wutong/${NAMESPACE}:${VERSION} -f Dockerfile
docker push swr.cn-southwest-2.myhuaweicloud.com/wutong/${NAMESPACE}:${VERSION}

# docker build . -t wutongpaas/${NAMESPACE}:${VERSION}
# docker push wutongpaas/${NAMESPACE}:${VERSION}
