package main

import (
	"blog/services/api-gateway/config"
	"blog/services/api-gateway/controller"
	"blog/services/api-gateway/middleware"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// 初始化配置
	cfg := config.LoadConfig()

	// 初始化控制器
	gatewayController := controller.NewGatewayController(cfg)

	// 初始化中间件
	corsMiddleware := middleware.NewCorsMiddleware()
	authMiddleware := middleware.NewAuthMiddleware()

	// 启动HTTP服务器
	server := controller.NewServer(cfg.Server.Port, gatewayController, corsMiddleware, authMiddleware)

	log.Printf("API Gateway started on port %s", cfg.Server.Port)

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("API Gateway shutting down...")
	server.Shutdown()
}
