package config

import (
	"blog/shared/config"
	"encoding/json"
	"log"
	"os"
)

// Config 钱包服务配置
type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	Redis    RedisConfig    `json:"redis"`
	Kafka    KafkaConfig    `json:"kafka"`
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

// LoadConfig 从Redis配置中心加载配置
func LoadConfig() *Config {
	// 尝试从Redis配置中心加载
	configCenter, err := config.NewConfigCenter("47.118.19.28:6379", "sta_go", 0)
	if err != nil {
		log.Printf("Failed to connect to Redis config center: %v, using default config", err)
		return loadDefaultConfig()
	}
	defer configCenter.Close()

	configData, err := configCenter.GetConfig("wallet-service")
	if err != nil {
		log.Printf("Failed to get config from Redis: %v, using default config", err)
		return loadDefaultConfig()
	}

	// 解析配置
	var cfg Config
	configBytes, err := json.Marshal(configData.Config)
	if err != nil {
		log.Printf("Failed to marshal config: %v, using default config", err)
		return loadDefaultConfig()
	}

	err = json.Unmarshal(configBytes, &cfg)
	if err != nil {
		log.Printf("Failed to unmarshal config: %v, using default config", err)
		return loadDefaultConfig()
	}

	log.Printf("Config loaded from Redis config center for wallet-service")
	return &cfg
}

// loadDefaultConfig 加载默认配置
func loadDefaultConfig() *Config {
	// 检查环境变量
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8002"
	}

	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "mysql"
	}

	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "3306"
	}

	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "redis"
	}

	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}

	kafkaHost := os.Getenv("KAFKA_HOST")
	if kafkaHost == "" {
		kafkaHost = "kafka"
	}

	kafkaPort := os.Getenv("KAFKA_PORT")
	if kafkaPort == "" {
		kafkaPort = "9092"
	}

	return &Config{
		Server: ServerConfig{
			Port: port,
		},
		Database: DatabaseConfig{
			DSN: "root:sta_go@tcp(" + dbHost + ":" + dbPort + ")/blog?charset=utf8mb4&parseTime=True&loc=Local",
		},
		Redis: RedisConfig{
			Addr:     redisHost + ":" + redisPort,
			Password: "sta_go",
			DB:       0,
		},
		Kafka: KafkaConfig{
			Brokers: []string{kafkaHost + ":" + kafkaPort},
		},
	}
}
