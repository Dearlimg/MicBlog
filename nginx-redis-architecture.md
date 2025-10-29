# Nginx + Redis 微服务架构使用指南

## 架构设计

```
客户端
   ↓
Nginx (负载均衡、反向代理)
   ↓
Redis (服务注册中心)
   ↓
┌─────────────────────────────────────┐
│  用户服务 (8001)                     │
│  钱包服务 (8002)                     │
│  评论服务 (8003)                     │
└─────────────────────────────────────┘
```

## 核心组件

### 1. Redis 作为服务注册中心

**优势**：
- ✅ 轻量级，资源占用少
- ✅ 支持 TTL（自动过期）
- ✅ 快速读写
- ✅ 易于部署

**服务注册**：
```go
// 服务启动时注册到Redis
registry.RegisterService(ServiceRegistration{
    ServiceName: "user-service",
    Address:     "127.0.0.1",
    Port:         8001,
    HealthURL:    "http://127.0.0.1:8001/health",
    Status:       "healthy",
})
```

**服务发现**：
```go
// 从Redis获取服务列表
services, err := registry.GetAllServices()
```

### 2. Nginx 作为反向代理

**配置位置**：`nginx.conf`

**功能**：
- ✅ 反向代理
- ✅ 负载均衡
- ✅ SSL终止
- ✅ 请求压缩
- ✅ CORS支持

## 快速开始

### 1. 安装依赖

```bash
# 安装Nginx
brew install nginx  # macOS
# 或 apt-get install nginx  # Ubuntu

# 安装Redis（如果还没有）
docker run -d --name redis -p 6379:6379 redis:7-alpine
```

### 2. 配置Nginx

```bash
# 复制配置文件
sudo cp nginx.conf /etc/nginx/sites-available/blog-gateway
sudo ln -s /etc/nginx/sites-available/blog-gateway /etc/nginx/sites-enabled/

# 测试配置
sudo nginx -t

# 启动Nginx
sudo nginx
```

### 3. 启动微服务

```bash
# 添加执行权限
chmod +x start-with-nginx.sh

# 启动所有服务
./start-with-nginx.sh
```

## 架构特点

### 服务注册流程

```
1. 服务启动
   ↓
2. 向Redis注册服务信息
   ↓
3. 定期发送心跳（每20秒）
   ↓
4. 服务下线时自动过期
```

### 请求路由流程

```
1. 客户端请求 → Nginx (80端口)
   ↓
2. Nginx根据路径路由
   ↓
3. 转发到对应的upstream池
   ↓
4. 负载均衡选择健康的服务实例
   ↓
5. 返回响应
```

### 服务发现流程

```
1. 配置更新器从Redis读取服务列表
   ↓
2. 生成Nginx upstream配置
   ↓
3. 重载Nginx配置
   ↓
4. Nginx使用新的服务列表
```

## API路由

所有请求都通过 Nginx 的 80 端口：

```bash
# 用户服务
POST http://localhost/api/v1/users/register
GET  http://localhost/api/v1/users/profile

# 钱包服务
GET  http://localhost/api/v1/wallets/123
POST http://localhost/api/v1/wallets/transfer

# 评论服务
GET  http://localhost/api/v1/comments/list
POST http://localhost/api/v1/comments/create
```

## 监控和日志

### 查看Nginx日志

```bash
# 访问日志
tail -f /var/log/nginx/gateway_access.log

# 错误日志
tail -f /var/log/nginx/gateway_error.log
```

### 查看微服务日志

```bash
# 用户服务
tail -f logs/user-service.log

# 钱包服务
tail -f logs/wallet-service.log

# 评论服务
tail -f logs/comment-service.log
```

### 查看Redis服务注册信息

```bash
# 连接到Redis
redis-cli -h 47.118.19.28 -p 6379 -a sta_go

# 查看所有服务
SMEMBERS services

# 查看特定服务
GET service:user-service
GET service:wallet-service
GET service:comment-service
```

## 服务健康检查

每个服务都会定期更新状态：

```go
// 每20秒发送一次心跳
ticker := time.NewTicker(20 * time.Second)
for range ticker.C {
    registry.RenewService(serviceName)
}
```

## 负载均衡策略

Nginx使用 `least_conn`（最少连接）策略：

```nginx
upstream user_service_pool {
    least_conn;  # 最少连接数
    server 127.0.0.1:8001;
    server 127.0.0.1:8011 backup;
}
```

## 故障转移

当服务不可用时：

1. **Nginx**：自动跳过不健康的服务
2. **Redis**：30秒后自动删除过期的服务
3. **配置更新器**：实时更新Nginx配置

## 优势

### 相比使用专用Consul：

1. **资源占用**：Redis比Consul更轻量
2. **学习成本**：Redis更常见，容易上手
3. **性能**：Redis读写性能极佳
4. **部署**：只需一个Redis容器

### 相比不使用注册中心：

1. **自动化**：服务自动注册和发现
2. **负载均衡**：多个实例自动分发请求
3. **健康检查**：自动过滤不健康的服务
4. **扩展性**：容易添加新实例

## 扩展部署

### 添加新服务实例

```bash
# 启动新的用户服务实例（端口8011）
REDIS_ADDRESS=47.118.19.28:6379 ./user-service -port 8011

# Nginx会自动检测并添加到upstream
```

### 扩容服务

只需要启动更多实例，注册中心会自动发现：

```bash
# 启动3个用户服务实例
REDIS_ADDRESS=47.118.19.28:6379 ./user-service &
REDIS_ADDRESS=47.118.19.28:6379 ./user-service -port 8011 &
REDIS_ADDRESS=47.118.19.28:6379 ./user-service -port 8012 &
```

## 故障排查

### 服务无法注册

```bash
# 检查Redis连接
redis-cli -h 47.118.19.28 -p 6379 -a sta_go ping
# 应该返回: PONG

# 检查服务是否注册
SMEMBERS services
```

### Nginx配置错误

```bash
# 测试配置
sudo nginx -t

# 查看错误日志
tail -f /var/log/nginx/error.log
```

### 服务无法访问

```bash
# 检查服务是否运行
ps aux | grep user-service

# 检查服务日志
tail -f logs/user-service.log

# 测试服务直连
curl http://localhost:8001/health
```

## 性能优化

### Nginx性能调优

```nginx
# worker进程数
worker_processes auto;

# 连接数
worker_connections 1024;

# 缓存
proxy_cache_path /var/cache/nginx levels=1:2 keys_zone=api_cache:10m;
proxy_cache api_cache;
```

### Redis性能优化

```conf
# redis.conf
maxmemory 256mb
maxmemory-policy allkeys-lru
```

## 总结

这是一个轻量级但功能完整的微服务架构：

- ✅ **Nginx**：高性能反向代理和负载均衡
- ✅ **Redis**：轻量级服务注册中心
- ✅ **自动服务发现**：无需手动配置
- ✅ **健康检查**：自动过滤不健康服务
- ✅ **负载均衡**：多实例自动分发请求
- ✅ **易于扩展**：添加实例只需启动新进程

适合中小型项目，资源占用少，易于部署和维护。
