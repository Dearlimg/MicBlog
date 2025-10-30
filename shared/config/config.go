package config

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

// ConfigCenter Redis配置中心
type ConfigCenter struct {
	client *redis.Client
	prefix string
}

// ConfigData 配置数据结构
type ConfigData struct {
	Service string                 `json:"service"`
	Config  map[string]interface{} `json:"config"`
	Version int64                  `json:"version"`
	Updated time.Time              `json:"updated"`
}

// NewConfigCenter 创建配置中心
func NewConfigCenter(addr, password string, db int) (*ConfigCenter, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}

	return &ConfigCenter{
		client: client,
		prefix: "config:",
	}, nil
}

// SetConfig 设置服务配置
func (cc *ConfigCenter) SetConfig(service string, config map[string]interface{}) error {
	configData := ConfigData{
		Service: service,
		Config:  config,
		Version: time.Now().Unix(),
		Updated: time.Now(),
	}

	data, err := json.Marshal(configData)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	key := cc.prefix + service
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = cc.client.Set(ctx, key, data, 0).Err()
	if err != nil {
		return fmt.Errorf("failed to set config: %v", err)
	}

	log.Printf("Config updated for service: %s", service)
	return nil
}

// GetConfig 获取服务配置
func (cc *ConfigCenter) GetConfig(service string) (*ConfigData, error) {
	key := cc.prefix + service
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	data, err := cc.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("config not found for service: %s", service)
		}
		return nil, fmt.Errorf("failed to get config: %v", err)
	}

	var configData ConfigData
	err = json.Unmarshal([]byte(data), &configData)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	return &configData, nil
}

// GetAllConfigs 获取所有配置
func (cc *ConfigCenter) GetAllConfigs() (map[string]*ConfigData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	keys, err := cc.client.Keys(ctx, cc.prefix+"*").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get keys: %v", err)
	}

	configs := make(map[string]*ConfigData)
	for _, key := range keys {
		service := key[len(cc.prefix):]
		config, err := cc.GetConfig(service)
		if err != nil {
			log.Printf("Failed to get config for service %s: %v", service, err)
			continue
		}
		configs[service] = config
	}

	return configs, nil
}

// DeleteConfig 删除服务配置
func (cc *ConfigCenter) DeleteConfig(service string) error {
	key := cc.prefix + service
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := cc.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete config: %v", err)
	}

	log.Printf("Config deleted for service: %s", service)
	return nil
}

// WatchConfig 监听配置变化
func (cc *ConfigCenter) WatchConfig(service string, callback func(*ConfigData)) error {
	key := cc.prefix + service
	ctx := context.Background()

	pubsub := cc.client.Subscribe(ctx, key)
	defer pubsub.Close()

	// 先获取当前配置
	config, err := cc.GetConfig(service)
	if err != nil {
		return fmt.Errorf("failed to get initial config: %v", err)
	}
	callback(config)

	// 监听变化
	for {
		msg, err := pubsub.ReceiveMessage(ctx)
		if err != nil {
			return fmt.Errorf("failed to receive message: %v", err)
		}

		var configData ConfigData
		err = json.Unmarshal([]byte(msg.Payload), &configData)
		if err != nil {
			log.Printf("Failed to unmarshal config update: %v", err)
			continue
		}

		callback(&configData)
	}
}

// Close 关闭配置中心
func (cc *ConfigCenter) Close() error {
	return cc.client.Close()
}

// InitializeDefaultConfigs 初始化默认配置
func (cc *ConfigCenter) InitializeDefaultConfigs() error {
	// API网关配置
	apiGatewayConfig := map[string]interface{}{
		"server": map[string]interface{}{
			"port": "8000",
		},
		"services": map[string]interface{}{
			"user_service": map[string]interface{}{
				"host": "localhost",
				"port": "8001",
			},
			"wallet_service": map[string]interface{}{
				"host": "localhost",
				"port": "8002",
			},
			"comment_service": map[string]interface{}{
				"host": "localhost",
				"port": "8003",
			},
			"shop_service": map[string]interface{}{
				"host": "localhost",
				"port": "8004",
			},
		},
	}

	// 用户服务配置
	userServiceConfig := map[string]interface{}{
		"server": map[string]interface{}{
			"port": "8001",
		},
		"database": map[string]interface{}{
			"dsn": "root:sta_go@tcp(47.118.19.28:3307)/blog?charset=utf8mb4&parseTime=True&loc=Local",
		},
		"redis": map[string]interface{}{
			"addr":     "47.118.19.28:6379",
			"password": "sta_go",
			"db":       0,
		},
		"kafka": map[string]interface{}{
			"brokers": []string{"localhost:9092"},
		},
		"email": map[string]interface{}{
			"host":     "smtp.qq.com",
			"port":     465,
			"username": "1492568061@qq.com",
			"password": "nbafgutnzsediibc",
			"from":     "1492568061@qq.com",
			"to":       []string{"1492568061@qq.com"},
			"is_ssl":   true,
		},
	}

	// 钱包服务配置
	walletServiceConfig := map[string]interface{}{
		"server": map[string]interface{}{
			"port": "8002",
		},
		"database": map[string]interface{}{
			"dsn": "root:sta_go@tcp(47.118.19.28:3307)/blog?charset=utf8mb4&parseTime=True&loc=Local",
		},
		"redis": map[string]interface{}{
			"addr":     "47.118.19.28:6379",
			"password": "sta_go",
			"db":       0,
		},
		"kafka": map[string]interface{}{
			"brokers": []string{"localhost:9092"},
		},
	}

	// 评论服务配置
	commentServiceConfig := map[string]interface{}{
		"server": map[string]interface{}{
			"port": "8003",
		},
		"database": map[string]interface{}{
			"dsn": "root:sta_go@tcp(47.118.19.28:3307)/blog?charset=utf8mb4&parseTime=True&loc=Local",
		},
		"redis": map[string]interface{}{
			"addr":     "47.118.19.28:6379",
			"password": "sta_go",
			"db":       0,
		},
		"kafka": map[string]interface{}{
			"brokers": []string{"localhost:9092"},
		},
	}

	// 商城服务配置
	shopServiceConfig := map[string]interface{}{
		"server": map[string]interface{}{
			"port": "8004",
		},
		"database": map[string]interface{}{
			"dsn": "root:sta_go@tcp(47.118.19.28:3307)/blog?charset=utf8mb4&parseTime=True&loc=Local",
		},
		"redis": map[string]interface{}{
			"addr":     "47.118.19.28:6379",
			"password": "sta_go",
			"db":       0,
		},
		"wallet": map[string]interface{}{
			"host": "wallet-service",
			"port": "8002",
		},
	}

	// 设置所有配置
	configs := map[string]map[string]interface{}{
		"api-gateway":     apiGatewayConfig,
		"user-service":    userServiceConfig,
		"wallet-service":  walletServiceConfig,
		"comment-service": commentServiceConfig,
		"shop-service":    shopServiceConfig,
	}

	for service, config := range configs {
		err := cc.SetConfig(service, config)
		if err != nil {
			return fmt.Errorf("failed to set config for %s: %v", service, err)
		}
	}

	log.Println("Default configurations initialized successfully")
	return nil
}
