package main

import (
	"blog/shared/kafka"
	"blog/wallet-service/config"
	"blog/wallet-service/controller"
	"blog/wallet-service/logic"
	"blog/wallet-service/repository"
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

	// 初始化Kafka生产者
	producer, err := kafka.NewProducer(cfg.Kafka.Brokers)
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}
	defer producer.Close()

	// 初始化Kafka消费者
	consumer, err := kafka.NewConsumer(cfg.Kafka.Brokers)
	if err != nil {
		log.Fatalf("Failed to create Kafka consumer: %v", err)
	}
	defer consumer.Close()

	// 初始化仓库
	walletRepo := repository.NewWalletRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)

	// 初始化业务逻辑
	walletLogic := logic.NewWalletLogic(walletRepo, transactionRepo, producer)

	// 初始化控制器
	walletController := controller.NewWalletController(walletLogic)

	// 启动HTTP服务器
	server := controller.NewServer(cfg.Server.Port, walletController)

	// 启动Kafka消费者
	go func() {
		err := consumer.ConsumeMessages(kafka.TopicWalletPayment, walletLogic.HandlePaymentEvent)
		if err != nil {
			log.Printf("Failed to consume messages: %v", err)
		}
	}()

	log.Printf("Wallet service starting on port %s", cfg.Server.Port)

	// 在goroutine中启动服务器
	go func() {
		if err := server.Start(); err != nil {
			log.Fatalf("Failed to start Wallet service: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Wallet service shutting down...")
	server.Shutdown()
}
