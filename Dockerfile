# Build the manager binary
FROM --platform=$BUILDPLATFORM golang:1.15 as builder
WORKDIR /workspace
# Copy the Go Modules manifests
COPY . .
ENV GOPROXY=https://goproxy.cn
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Build
ARG TARGETOS TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GO111MODULE=on go build -o /cloud-adaptor ./cmd/cloud-adaptor

FROM --platform=$BUILDPLATFORM wutongpaas/alpine:3.15
ARG TARGETOS TARGETARCH TARGETPLATFORM
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

EXPOSE 8080

ENTRYPOINT ["./cloud-adaptor"]