package main

import (
	"blog/comment-service/config"
	"blog/comment-service/controller"
	"blog/comment-service/logic"
	"blog/comment-service/repository"
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
	commentRepo := repository.NewCommentRepository(db)

	// 初始化业务逻辑
	commentLogic := logic.NewCommentLogic(commentRepo, producer)

	// 初始化控制器
	commentController := controller.NewCommentController(commentLogic)

	// 启动HTTP服务器
	server := controller.NewServer(cfg.Server.Port, commentController)

	// 启动Kafka消费者
	go func() {
		err := consumer.ConsumeMessages(kafka.TopicCommentCreate, commentLogic.HandleCommentEvent)
		if err != nil {
			log.Printf("Failed to consume messages: %v", err)
		}
	}()

	log.Printf("Comment service started on port %s", cfg.Server.Port)

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Comment service shutting down...")
	server.Shutdown()
}
