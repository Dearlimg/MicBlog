FROM golang:1.23-alpine AS builder

WORKDIR /app

# 复制go mod文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建API网关
RUN CGO_ENABLED=0 GOOS=linux go build -o api-gateway ./services/api-gateway

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# 复制构建的二进制文件
COPY --from=builder /app/api-gateway .

# 暴露端口
EXPOSE 8000

# 运行服务
CMD ["./api-gateway"]
