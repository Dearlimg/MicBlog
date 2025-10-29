package kafka

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/Shopify/sarama"
)

// Producer Kafka生产者
type Producer struct {
	producer sarama.SyncProducer
}

// Consumer Kafka消费者
type Consumer struct {
	consumer sarama.Consumer
}

// Message Kafka消息结构
type Message struct {
	Topic     string      `json:"topic"`
	Key       string      `json:"key"`
	Value     interface{} `json:"value"`
	Timestamp time.Time   `json:"timestamp"`
}

// NewProducer 创建Kafka生产者
func NewProducer(brokers []string) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %v", err)
	}

	return &Producer{producer: producer}, nil
}

// SendMessage 发送消息
func (p *Producer) SendMessage(topic, key string, value interface{}) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.StringEncoder(jsonData),
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %v", err)
	}

	log.Printf("Message sent to topic %s, partition %d, offset %d", topic, partition, offset)
	return nil
}

// Close 关闭生产者
func (p *Producer) Close() error {
	return p.producer.Close()
}

// NewConsumer 创建Kafka消费者
func NewConsumer(brokers []string) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %v", err)
	}

	return &Consumer{consumer: consumer}, nil
}

// ConsumeMessages 消费消息
func (c *Consumer) ConsumeMessages(topic string, handler func(*sarama.ConsumerMessage) error) error {
	partitionList, err := c.consumer.Partitions(topic)
	if err != nil {
		return fmt.Errorf("failed to get partitions: %v", err)
	}

	for _, partition := range partitionList {
		pc, err := c.consumer.ConsumePartition(topic, partition, sarama.OffsetNewest)
		if err != nil {
			return fmt.Errorf("failed to consume partition: %v", err)
		}

		go func(pc sarama.PartitionConsumer) {
			defer pc.AsyncClose()
			for {
				select {
				case msg := <-pc.Messages():
					if err := handler(msg); err != nil {
						log.Printf("Error handling message: %v", err)
					}
				case err := <-pc.Errors():
					log.Printf("Consumer error: %v", err)
				}
			}
		}(pc)
	}

	return nil
}

// Close 关闭消费者
func (c *Consumer) Close() error {
	return c.consumer.Close()
}

// Topics 定义Kafka主题
const (
	TopicUserRegister    = "user.register"
	TopicUserLogin       = "user.login"
	TopicUserEmailVerify = "user.email.verify"
	TopicWalletPayment   = "wallet.payment"
	TopicCommentCreate   = "comment.create"
	TopicCommentUpdate   = "comment.update"
	TopicCommentDelete   = "comment.delete"
)
