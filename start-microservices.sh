#!/bin/bash

# 微服务启动脚本

echo "🚀 启动博客微服务系统..."

# 检查Docker是否运行
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker未运行，请先启动Docker"
    exit 1
fi

# 停止现有容器
echo "🛑 停止现有容器..."
docker-compose down

# 构建并启动服务
echo "🔨 构建并启动服务..."
docker-compose up --build -d

# 等待服务启动
echo "⏳ 等待服务启动..."
sleep 30

# 检查服务状态
echo "📊 检查服务状态..."
docker-compose ps

# 显示服务日志
echo "📝 显示服务日志..."
docker-compose logs --tail=20

echo "✅ 微服务系统启动完成！"
echo ""
echo "🌐 服务访问地址："
echo "  API网关: http://localhost:8000"
echo "  用户服务: http://localhost:8001"
echo "  钱包服务: http://localhost:8002"
echo "  评论服务: http://localhost:8003"
echo ""
echo "📋 常用命令："
echo "  查看日志: docker-compose logs -f [service_name]"
echo "  停止服务: docker-compose down"
echo "  重启服务: docker-compose restart [service_name]"
