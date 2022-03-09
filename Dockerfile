FROM golang:1.13 as builder
ENV CGO_ENABLED=0
ENV GOPATH=/go
ENV GOPROXY=https://goproxy.cn

WORKDIR /app
COPY . .

ARG LDFLAGS
RUN GOOS=linux GO111MODULE=on go build -ldflags "$LDFLAGS" -o /cloud-adaptor ./cmd/cloud-adaptor

FROM alpine/helm:3.8.0
WORKDIR /app
RUN apk add --update tzdata && \
    apk add --update apache2-utils && \
    rm -rf /var/cache/apk/* && \
    mkdir /app/data
RUN helm repo add wutong https://wutong-paas.github.io/helm-charts && helm repo update
ENV TZ=Asia/Shanghai
ENV DB_PATH=/app/data/cloudadaptor
ENV CONFIG_DIR=/app/data/cloudadaptor
ENV HELM_PATH=/usr/local/bin/helm
ENV MYSQL_DB=console

COPY --from=builder /cloud-adaptor .

VOLUME /app/data

ENTRYPOINT ["./cloud-adaptor"]