package main

import (
	"blog/shared/kafka"
	"blog/user-service/config"
	"blog/user-service/controller"
	"blog/user-service/logic"
	"blog/user-service/repository"
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

	// 初始化Kafka生产者（可选）
	var producer *kafka.Producer
	var consumer *kafka.Consumer
	if len(cfg.Kafka.Brokers) > 0 && cfg.Kafka.Brokers[0] != "" {
		var err error
		producer, err = kafka.NewProducer(cfg.Kafka.Brokers)
		if err != nil {
			log.Printf("Failed to create Kafka producer: %v, continuing without Kafka", err)
		} else {
			defer producer.Close()
		}

		// 初始化Kafka消费者（可选）
		consumer, err = kafka.NewConsumer(cfg.Kafka.Brokers)
		if err != nil {
			log.Printf("Failed to create Kafka consumer: %v, continuing without Kafka", err)
		} else {
			defer consumer.Close()
		}
	} else {
		log.Println("Kafka not configured, running without Kafka")
	}

	// 初始化仓库
	userRepo := repository.NewUserRepository(db)
	emailRepo := repository.NewEmailRepository(redisClient)

	// 初始化业务逻辑
	userLogic := logic.NewUserLogic(userRepo, emailRepo, producer, cfg.Email, cfg.JWT.Secret)

	// 初始化控制器
	userController := controller.NewUserController(userLogic)

	// 启动HTTP服务器
	server := controller.NewServer(cfg.Server.Port, userController)

	// 启动Kafka消费者（如果可用）
	if consumer != nil {
		go func() {
			err := consumer.ConsumeMessages(kafka.TopicUserEmailVerify, userLogic.HandleEmailVerification)
			if err != nil {
				log.Printf("Failed to consume messages: %v", err)
			}
		}()
	}

	log.Printf("User service starting on port %s", cfg.Server.Port)

	// 在goroutine中启动服务器
	go func() {
		if err := server.Start(); err != nil {
			log.Fatalf("Failed to start User service: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("User service shutting down...")
	server.Shutdown()
}
