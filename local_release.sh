go mod download
go build -o ./bin/cloud-adaptor ./cmd/cloud-adaptor
docker build . -t swr.cn-southwest-2.myhuaweicloud.com/wutong/wutong-adaptor:v1.0.0-stable
docker push swr.cn-southwest-2.myhuaweicloud.com/wutong/wutong-adaptor:v1.0.0-stable