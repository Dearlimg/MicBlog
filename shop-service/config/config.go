package config

import (
	"blog/shared/config"
	"encoding/json"
	"log"
	"os"
)

// Config 商城服务配置
type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	Redis    RedisConfig    `json:"redis"`
	Wallet   WalletConfig   `json:"wallet"`
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

// WalletConfig 钱包服务配置
type WalletConfig struct {
	Host string `json:"host"`
	Port string `json:"port"`
}

// LoadConfig 从Redis配置中心加载配置
func LoadConfig() *Config {
	// 尝试从Redis配置中心加载
	configCenter, err := config.NewConfigCenter("47.118.19.28:6379", "sta_go", 0)
	if err != nil {
		log.Printf("Failed to connect to Redis config center: %v, using default config", err)
		return LoadDefaultConfig()
	}
	defer configCenter.Close()

	configData, err := configCenter.GetConfig("shop-service")
	if err != nil {
		log.Printf("Failed to get config from Redis: %v, using default config", err)
		return LoadDefaultConfig()
	}

	// 解析配置
	var cfg Config
	configBytes, err := json.Marshal(configData.Config)
	if err != nil {
		log.Printf("Failed to marshal config: %v, using default config", err)
		return LoadDefaultConfig()
	}

	err = json.Unmarshal(configBytes, &cfg)
	if err != nil {
		log.Printf("Failed to unmarshal config: %v, using default config", err)
		return LoadDefaultConfig()
	}

	log.Printf("Config loaded from Redis config center for shop-service")
	return &cfg
}

// LoadDefaultConfig 加载默认配置
func LoadDefaultConfig() *Config {
	// 检查环境变量
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8004"
	}

	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "47.118.19.28"
	}

	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "3307"
	}

	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "47.118.19.28"
	}

	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}

	walletHost := os.Getenv("WALLET_SERVICE_HOST")
	if walletHost == "" {
		walletHost = "localhost"
	}

	walletPort := os.Getenv("WALLET_SERVICE_PORT")
	if walletPort == "" {
		walletPort = "8002"
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
		Wallet: WalletConfig{
			Host: walletHost,
			Port: walletPort,
		},
	}
}
