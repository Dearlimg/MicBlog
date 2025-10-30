package controller

import (
	"blog/api-gateway/config"
	"blog/api-gateway/middleware"
	"bytes"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/Dearlimg/Goutils/pkg/app"
	"github.com/gin-gonic/gin"
)

// GatewayController API网关控制器
type GatewayController struct {
	config *config.Config
	client *http.Client
}

// NewGatewayController 创建API网关控制器
func NewGatewayController(cfg *config.Config) *GatewayController {
	return &GatewayController{
		config: cfg,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ProxyRequest 代理请求到微服务
func (gc *GatewayController) ProxyRequest(c *gin.Context) {
	// 获取目标服务
	targetService := c.Param("service")
	targetPath := c.Param("path")

	// 构建目标URL
	var targetURL string
	switch targetService {
	case "users":
		targetURL = "http://" + gc.config.Services.UserService.Host + ":" + gc.config.Services.UserService.Port + "/api/v1/users" + targetPath
	case "wallets":
		targetURL = "http://" + gc.config.Services.WalletService.Host + ":" + gc.config.Services.WalletService.Port + "/api/v1/wallets" + targetPath
	case "comments":
		targetURL = "http://" + gc.config.Services.CommentService.Host + ":" + gc.config.Services.CommentService.Port + "/api/v1/comments" + targetPath
	case "products", "orders", "shop":
		targetURL = "http://" + gc.config.Services.ShopService.Host + ":" + gc.config.Services.ShopService.Port + "/api/v1" + targetPath
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unknown service"})
		return
	}

	// 读取请求体
	var bodyBytes []byte
	if c.Request.Body != nil {
		bodyBytes, _ = io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	// 创建代理请求
	req, err := http.NewRequest(c.Request.Method, targetURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	// 复制请求头
	for key, values := range c.Request.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// 发送请求
	resp, err := gc.client.Do(req)
	if err != nil {
		log.Printf("Failed to proxy request to %s: %v", targetURL, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to proxy request"})
		return
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
		return
	}

	// 复制响应头
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	// 返回响应
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
}

// HealthCheck 健康检查
func (gc *GatewayController) HealthCheck(c *gin.Context) {
	rly := app.NewResponse(c)

	// 检查各个微服务的健康状态
	services := map[string]string{
		"user-service":    "http://" + gc.config.Services.UserService.Host + ":" + gc.config.Services.UserService.Port + "/health",
		"wallet-service":  "http://" + gc.config.Services.WalletService.Host + ":" + gc.config.Services.WalletService.Port + "/health",
		"comment-service": "http://" + gc.config.Services.CommentService.Host + ":" + gc.config.Services.CommentService.Port + "/health",
		"shop-service":    "http://" + gc.config.Services.ShopService.Host + ":" + gc.config.Services.ShopService.Port + "/api/v1/health",
	}

	status := make(map[string]string)
	for serviceName, healthURL := range services {
		resp, err := gc.client.Get(healthURL)
		if err != nil || resp.StatusCode != http.StatusOK {
			status[serviceName] = "unhealthy"
		} else {
			status[serviceName] = "healthy"
		}
		if resp != nil {
			resp.Body.Close()
		}
	}

	rly.Reply(nil, gin.H{
		"gateway":  "healthy",
		"services": status,
	})
}

// Server HTTP服务器
type Server struct {
	router *gin.Engine
	port   string
}

// NewServer 创建HTTP服务器
func NewServer(port string, gatewayController *GatewayController, corsMiddleware *middleware.CorsMiddleware, authMiddleware *middleware.AuthMiddleware) *Server {
	router := gin.New()
	router.Use(gin.Recovery(), gin.Logger())

	// 添加中间件
	router.Use(corsMiddleware.Handle())
	router.Use(authMiddleware.Handle())

	// API路由
	api := router.Group("/api/v1")
	{
		// 健康检查
		api.GET("/health", gatewayController.HealthCheck)

		// 代理路由
		api.Any("/:service/*path", gatewayController.ProxyRequest)
	}

	return &Server{
		router: router,
		port:   port,
	}
}

// Start 启动服务器
func (s *Server) Start() error {
	return s.router.Run(":" + s.port)
}

// Shutdown 关闭服务器
func (s *Server) Shutdown() {
	// 这里可以添加优雅关闭逻辑
}
