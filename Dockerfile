FROM --platform=$BUILDPLATFORM golang:1.13 as builder
ENV CGO_ENABLED=0
ENV GOPATH=/go
ENV GOPROXY=https://goproxy.io

WORKDIR /app
COPY . .

ARG LDFLAGS
ARG TARGETOS TARGETARCH
RUN GO111MODULE=on GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags "$LDFLAGS" -o /cloud-adaptor ./cmd/cloud-adaptor

FROM --platform=$BUILDPLATFORM wutongpaas/alpine:3.15
ARG TARGETPLATFORM
WORKDIR /app
RUN apk add --update apache2-utils && \
    rm -rf /var/cache/apk/* && \
    mkdir /app/data && \
    if [ "${TARGETPLATFORM}" = "linux/arm64" ]; then \
      wget https://wutong-paas-public.obs.cn-east-3.myhuaweicloud.com/offline/helm-arm64 && chmod +x helm-arm64 && mv helm-arm64 /usr/local/bin/helm; \
    else \
      wget https://wutong-paas-public.obs.cn-east-3.myhuaweicloud.com/offline/helm && chmod +x helm && mv helm /usr/local/bin/helm; \
    fi && \
    helm repo add wutong https://wutong-paas.github.io/helm-charts && helm repo update
ENV TZ=Asia/Shanghai
ENV DB_PATH=/app/data/cloudadaptor
ENV CONFIG_DIR=/app/data/cloudadaptor
ENV HELM_PATH=/usr/local/bin/helm
ENV MYSQL_DB=console

COPY --from=builder /cloud-adaptor .

VOLUME /app/data

ENTRYPOINT ["./cloud-adaptor"]