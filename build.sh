#! /bin/bash
# References:
# 1. https://docs.docker.com/buildx/working-with-buildx/
# 2. https://docs.docker.com/engine/reference/commandline/buildx/

# Requirements:
# 1. docker run --privileged --rm tonistiigi/binfmt --install all
# 2. docker login dockerhub
# 3. docker login myhuaweicloud-swr

export VERSION=v1.0.0-stable
docker buildx create --use --name adaptorbuilder
docker buildx build --platform linux/amd64,linux/arm64 --push -t wutongpaas/wutong-adaptor:${VERSION} -f Dockerfile . 
docker buildx build --platform linux/amd64,linux/arm64 --push -t swr.cn-southwest-2.myhuaweicloud.com/wutong/wutong-adaptor:${VERSION} -f Dockerfile . 
docker buildx rm adaptorbuilder