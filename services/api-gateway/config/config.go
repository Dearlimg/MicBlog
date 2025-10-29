package config

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
}

// ServiceConfig 服务配置
type ServiceConfig struct {
	Host string `json:"host"`
	Port string `json:"port"`
}

// LoadConfig 加载配置
func LoadConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port: "8000",
		},
		Services: ServicesConfig{
			UserService: ServiceConfig{
				Host: "localhost",
				Port: "8001",
			},
			WalletService: ServiceConfig{
				Host: "localhost",
				Port: "8002",
			},
			CommentService: ServiceConfig{
				Host: "localhost",
				Port: "8003",
			},
		},
	}
}
