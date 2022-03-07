# FROM golang:1.13 as builder
# ENV CGO_ENABLED=0
# ENV GOPATH=/go
# ENV GOPROXY=https://goproxy.cn

WORKDIR /app
COPY . .

ARG GOARCH
ARG LDFLAGS
RUN go build -ldflags "$LDFLAGS" -o /cloud-adaptor ./cmd/cloud-adaptor

FROM alpine:3.11.2
WORKDIR /app
RUN apk add --update tzdata \
    && apk add --update apache2-utils \
    && rm -rf /var/cache/apk/* \
    && mkdir /app/data \
    && wget -q https://wutong-paas-public.obs.cn-east-3.myhuaweicloud.com/offline/helm \
    && chmod +x helm \
    && mv helm /usr/local/bin \
    && helm add wutong https://wutong-paas.github.io/helm-charts && helm update \
ENV TZ=Asia/Shanghai
ENV DB_PATH=/app/data/cloudadaptor
ENV CHART_PATH=/app/chart
ENV CONFIG_DIR=/app/data/cloudadaptor
ENV HELM_PATH=/usr/local/bin/helm
ENV MYSQL_DB=console

COPY --from=builder /cloud-adaptor .
COPY --from=builder /app/chart ./chart

VOLUME /app/data

ENTRYPOINT ["./cloud-adaptor"]