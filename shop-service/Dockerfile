FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

# 构建商城服务
RUN CGO_ENABLED=0 GOOS=linux go build -o shop-service ./shop-service

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/shop-service .

CMD ["./shop-service"]

