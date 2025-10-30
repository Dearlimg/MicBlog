package controller

import (
	"github.com/gin-gonic/gin"
)

// Server HTTP服务器
type Server struct {
	router *gin.Engine
	port   string
}

// NewServer 创建HTTP服务器
func NewServer(port string, productController *ProductController, orderController *OrderController, cartController *CartController) *Server {
	router := gin.New()
	router.Use(gin.Recovery(), gin.Logger())

	// 健康检查路由
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy", "service": "shop-service"})
	})

	// API路由
	api := router.Group("/api/v1")
	{
		// 商品相关路由
		products := api.Group("/products")
		{
			products.POST("", productController.CreateProduct)
			products.GET("", productController.GetProducts)
			products.GET("/:id", productController.GetProduct)
			products.PUT("/:id", productController.UpdateProduct)
			products.DELETE("/:id", productController.DeleteProduct)
		}

		// 订单相关路由
		orders := api.Group("/orders")
		{
			orders.POST("", orderController.CreateOrder)
			orders.GET("/:id", orderController.GetOrder)
			orders.PUT("/:id/cancel", orderController.CancelOrder)
		}

		ordersUser := api.Group("/users/:user_id/orders")
		{
			ordersUser.GET("", orderController.GetUserOrders)
		}

		// 购物车相关路由
		cart := api.Group("/users/:user_id/cart")
		{
			cart.POST("", cartController.AddToCart)
			cart.GET("", cartController.GetCart)
			cart.PUT("/:product_id", cartController.UpdateCartItem)
			cart.DELETE("/:product_id", cartController.RemoveFromCart)
			cart.DELETE("", cartController.ClearCart)
		}
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
