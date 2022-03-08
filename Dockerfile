FROM golang:1.13 as builder
ENV CGO_ENABLED=0
ENV GOPATH=/go
ENV GOPROXY=https://goproxy.cn

WORKDIR /app
COPY . .

ARG GOARCH=amd64
ARG LDFLAGS
RUN GOOS=linux GOARCH="${GOARCH}" GO111MODULE=on go build -ldflags "$LDFLAGS" -o /cloud-adaptor ./cmd/cloud-adaptor

FROM alpine:3.11.2
WORKDIR /app
RUN apk add --update tzdata && \
    apk add --update apache2-utils && \
    rm -rf /var/cache/apk/* && \
    mkdir /app/data && \
    if ["$[GOARCH]" = "arm64" ]; then \
      wget https://wutong-paas-public.obs.cn-east-3.myhuaweicloud.com/offline/helm64 && chmod +x helm64 && mv helm64 /usr/local/bin/helm; \
    else \
      wget https://wutong-paas-public.obs.cn-east-3.myhuaweicloud.com/offline/helm && chmod +x helm && mv helm /usr/local/bin/; \
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