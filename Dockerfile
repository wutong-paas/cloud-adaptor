FROM wutongpaas/alpine:3.15
WORKDIR /app
RUN apk add --update apache2-utils && \
    rm -rf /var/cache/apk/* && \
    mkdir /app/data && \
    wget https://wutong-paas.obs.cn-east-3.myhuaweicloud.com/amd/helm && chmod +x helm && mv helm /usr/local/bin/helm && \
    helm repo add wutong https://wutong-paas.github.io/helm-charts && helm repo update
ENV TZ=Asia/Shanghai
ENV DB_PATH=/app/data/cloudadaptor
ENV CONFIG_DIR=/app/data/cloudadaptor
ENV HELM_PATH=/usr/local/bin/helm
ENV MYSQL_DB=console

COPY ./bin/cloud-adaptor .

VOLUME /app/data

EXPOSE 8080

ENTRYPOINT ["./cloud-adaptor"]