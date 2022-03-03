FROM golang:1.13 as builder
ENV CGO_ENABLED=0
ENV GOPATH=/go
ENV GOPROXY=https://goproxy.cn

WORKDIR /app
COPY . .

ARG GOARCH
ARG LDFLAGS
RUN go build -ldflags "$LDFLAGS" -o /cloud-adaptor ./cmd/cloud-adaptor

FROM alpine:3.11.2
WORKDIR /app
RUN apk add --update tzdata \
    && apk add --update apache2-utils \
    && rm -rf /var/cache/apk/*
ENV TZ=Asia/Shanghai
COPY --from=builder cloud-adaptor .

ENTRYPOINT ["./cloud-adaptor"]
