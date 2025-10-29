package main

import (
	"blog/shop-service/config"
	"blog/shop-service/controller"
	"blog/shop-service/logic"
	"blog/shop-service/repository"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// 初始化配置
	cfg := config.LoadConfig()

	// 初始化数据库
	db := repository.InitDB(cfg.Database.DSN)

	// 初始化仓库
	productRepo := repository.NewProductRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	cartRepo := repository.NewCartRepository(db)

	// 构建钱包服务URL（用于HTTP调用）
	walletURL := cfg.Wallet.Host + ":" + cfg.Wallet.Port

	// 初始化业务逻辑
	productLogic := logic.NewProductLogic(productRepo)
	orderLogic := logic.NewOrderLogic(productRepo, orderRepo, cartRepo, walletURL)
	cartLogic := logic.NewCartLogic(cartRepo, productRepo)

	// 初始化控制器
	productController := controller.NewProductController(productLogic)
	orderController := controller.NewOrderController(orderLogic)
	cartController := controller.NewCartController(cartLogic)

	// 启动HTTP服务器
	server := controller.NewServer(cfg.Server.Port, productController, orderController, cartController)

	log.Printf("Shop service started on port %s", cfg.Server.Port)

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shop service shutting down...")
	server.Shutdown()
}
