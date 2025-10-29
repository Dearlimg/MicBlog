package config

import (
	"blog/shared/config"
	"encoding/json"
	"log"
	"os"
)

// Config API网关配置
type Config struct {
	Server   ServerConfig   `json:"server"`
	Services ServicesConfig `json:"services"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port string `json:"port"`
}

// ServicesConfig 微服务配置
type ServicesConfig struct {
	UserService    ServiceConfig `json:"user_service"`
	WalletService  ServiceConfig `json:"wallet_service"`
	CommentService ServiceConfig `json:"comment_service"`
	ShopService    ServiceConfig `json:"shop_service"`
}

// ServiceConfig 服务配置
type ServiceConfig struct {
	Host string `json:"host"`
	Port string `json:"port"`
}

// LoadConfig 从Redis配置中心加载配置
func LoadConfig() *Config {
	// 尝试从Redis配置中心加载
	configCenter, err := config.NewConfigCenter("redis:6379", "sta_go", 0)
	if err != nil {
		log.Printf("Failed to connect to Redis config center: %v, using default config", err)
		return loadDefaultConfig()
	}
	defer configCenter.Close()

	configData, err := configCenter.GetConfig("api-gateway")
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

	log.Printf("Config loaded from Redis config center for api-gateway")
	return &cfg
}

// loadDefaultConfig 加载默认配置
func loadDefaultConfig() *Config {
	// 检查环境变量
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8000"
	}

	userHost := os.Getenv("USER_SERVICE_HOST")
	if userHost == "" {
		userHost = "user-service"
	}

	walletHost := os.Getenv("WALLET_SERVICE_HOST")
	if walletHost == "" {
		walletHost = "wallet-service"
	}

	commentHost := os.Getenv("COMMENT_SERVICE_HOST")
	if commentHost == "" {
		commentHost = "comment-service"
	}

	shopHost := os.Getenv("SHOP_SERVICE_HOST")
	if shopHost == "" {
		shopHost = "shop-service"
	}

	return &Config{
		Server: ServerConfig{
			Port: port,
		},
		Services: ServicesConfig{
			UserService: ServiceConfig{
				Host: userHost,
				Port: "8001",
			},
			WalletService: ServiceConfig{
				Host: walletHost,
				Port: "8002",
			},
			CommentService: ServiceConfig{
				Host: commentHost,
				Port: "8003",
			},
			ShopService: ServiceConfig{
				Host: shopHost,
				Port: "8004",
			},
		},
	}
}
