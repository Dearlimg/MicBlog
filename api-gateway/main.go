package main

import (
	"blog/api-gateway/config"
	"blog/api-gateway/controller"
	"blog/api-gateway/middleware"
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

	log.Printf("API Gateway starting on port %s", cfg.Server.Port)

	// 在goroutine中启动服务器
	go func() {
		if err := server.Start(); err != nil {
			log.Fatalf("Failed to start API Gateway: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("API Gateway shutting down...")
	server.Shutdown()
}
