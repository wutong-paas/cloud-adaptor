# Build the manager binary
FROM golang:1.21 as builder
WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
ENV GOPROXY=https://goproxy.cn,direct
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY api/ api/
COPY cmd/ cmd/
COPY docs/ docs/
COPY internal/ internal/
COPY pkg/ pkg/
COPY version/ version/
COPY tools.go tools.go

# Build
ARG TARGETOS TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GO111MODULE=on go build -a -o /cloud-adaptor ./cmd/cloud-adaptor

FROM swr.cn-southwest-2.myhuaweicloud.com/wutong/alpine:3.15
ARG TARGETARCH
WORKDIR /app
RUN apk add --update apache2-utils && \
    rm -rf /var/cache/apk/* && \
    mkdir /app/data && \
    if [ "${TARGETARCH}" = "arm64" ]; then \
    wget https://wutong-paas.obs.cn-east-3.myhuaweicloud.com/arm/helm-arm64 && chmod +x helm-arm64 && mv helm-arm64 /usr/local/bin/helm; \
    else \
    wget https://wutong-paas.obs.cn-east-3.myhuaweicloud.com/amd/helm && chmod +x helm && mv helm /usr/local/bin/helm; \
    fi && \
    helm repo add wutong https://wutong-paas.github.io/helm-charts && helm repo update && \
    helm pull wutong/wutong-operator --version 1.5.0 --untar --untardir /app/charts
ENV TZ=Asia/Shanghai
ENV CONFIG_DIR=/app/data/cloudadaptor
ENV HELM_PATH=/usr/local/bin/helm
ENV MYSQL_DB=console

COPY --from=builder /cloud-adaptor .

VOLUME /app/data

EXPOSE 8080

ENTRYPOINT ["./cloud-adaptor"]
