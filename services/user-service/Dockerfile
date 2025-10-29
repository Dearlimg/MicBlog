FROM golang:1.23-alpine AS builder

WORKDIR /app

# 复制go mod文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建用户服务
RUN CGO_ENABLED=0 GOOS=linux go build -o user-service ./services/user-service

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# 复制构建的二进制文件
COPY --from=builder /app/user-service .

# 暴露端口
EXPOSE 8001

# 运行服务
CMD ["./user-service"]
