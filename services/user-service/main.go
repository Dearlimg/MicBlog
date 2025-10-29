package main

import (
	"blog/services/user-service/config"
	"blog/services/user-service/controller"
	"blog/services/user-service/logic"
	"blog/services/user-service/repository"
	"blog/shared/kafka"
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

	// 初始化Redis
	redisClient := repository.InitRedis(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)

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
	userRepo := repository.NewUserRepository(db)
	emailRepo := repository.NewEmailRepository(redisClient)

	// 初始化业务逻辑
	userLogic := logic.NewUserLogic(userRepo, emailRepo, producer, cfg.Email)

	// 初始化控制器
	userController := controller.NewUserController(userLogic)

	// 启动HTTP服务器
	server := controller.NewServer(cfg.Server.Port, userController)

	// 启动Kafka消费者
	go func() {
		err := consumer.ConsumeMessages(kafka.TopicUserEmailVerify, userLogic.HandleEmailVerification)
		if err != nil {
			log.Printf("Failed to consume messages: %v", err)
		}
	}()

	log.Printf("User service started on port %s", cfg.Server.Port)

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("User service shutting down...")
	server.Shutdown()
}
