package config

import (
	"blog/shared/email"
)

// Config 用户服务配置
type Config struct {
	Server   ServerConfig      `json:"server"`
	Database DatabaseConfig    `json:"database"`
	Redis    RedisConfig       `json:"redis"`
	Kafka    KafkaConfig       `json:"kafka"`
	Email    email.EmailConfig `json:"email"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port string `json:"port"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	DSN string `json:"dsn"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Addr     string `json:"addr"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

// KafkaConfig Kafka配置
type KafkaConfig struct {
	Brokers []string `json:"brokers"`
}

// LoadConfig 加载配置
func LoadConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port: "8001",
		},
		Database: DatabaseConfig{
			DSN: "root:sta_go@tcp(47.118.19.28:3307)/blog?charset=utf8mb4&parseTime=True&loc=Local",
		},
		Redis: RedisConfig{
			Addr:     "47.118.19.28:6379",
			Password: "sta_go",
			DB:       0,
		},
		Kafka: KafkaConfig{
			Brokers: []string{"47.118.19.28:9092"},
		},
		Email: email.EmailConfig{
			Host:     "smtp.qq.com",
			Port:     465,
			Username: "1492568061@qq.com",
			Password: "nbafgutnzsediibc",
			From:     "1492568061@qq.com",
			To:       []string{"1492568061@qq.com"},
			IsSSL:    true,
		},
	}
}
