package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

// ServiceRegistration 服务注册信息
type ServiceRegistration struct {
	ServiceName string    `json:"service_name"`
	Address     string    `json:"address"`
	Port        int       `json:"port"`
	HealthURL   string    `json:"health_url"`
	Status      string    `json:"status"` // healthy, unhealthy
	LastCheck   time.Time `json:"last_check"`
}

// ServiceRegistry Redis服务注册中心
type ServiceRegistry struct {
	client  *redis.Client
	ctx     context.Context
	service *ServiceRegistration
}

// NewServiceRegistry 创建服务注册中心
func NewServiceRegistry(redisAddr, redisPassword string, db int) (*ServiceRegistry, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       db,
	})

	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}

	return &ServiceRegistry{
		client: rdb,
		ctx:    ctx,
	}, nil
}

// RegisterService 注册服务
func (sr *ServiceRegistry) RegisterService(service ServiceRegistration) error {
	key := fmt.Sprintf("service:%s", service.ServiceName)

	service.LastCheck = time.Now()
	serviceData, err := json.Marshal(service)
	if err != nil {
		return err
	}

	// 设置过期时间为30秒，服务需要定期续期
	err = sr.client.Set(sr.ctx, key, serviceData, 30*time.Second).Err()
	if err != nil {
		return err
	}

	// 添加到服务列表
	sr.client.SAdd(sr.ctx, "services", service.ServiceName)

	log.Printf("✅ 服务 %s 已注册到Redis: %s:%d", service.ServiceName, service.Address, service.Port)
	return nil
}

// GetService 获取服务信息
func (sr *ServiceRegistry) GetService(serviceName string) (*ServiceRegistration, error) {
	key := fmt.Sprintf("service:%s", serviceName)
	val, err := sr.client.Get(sr.ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("service not found: %s", serviceName)
	}

	var service ServiceRegistration
	err = json.Unmarshal([]byte(val), &service)
	if err != nil {
		return nil, err
	}

	return &service, nil
}

// GetAllServices 获取所有服务
func (sr *ServiceRegistry) GetAllServices() ([]ServiceRegistration, error) {
	// 获取所有服务名称
	serviceNames, err := sr.client.SMembers(sr.ctx, "services").Result()
	if err != nil {
		return nil, err
	}

	var services []ServiceRegistration
	for _, name := range serviceNames {
		service, err := sr.GetService(name)
		if err != nil {
			continue
		}
		services = append(services, *service)
	}

	return services, nil
}

// RenewService 续期服务（心跳）
func (sr *ServiceRegistry) RenewService(serviceName string) error {
	service, err := sr.GetService(serviceName)
	if err != nil {
		return err
	}

	service.LastCheck = time.Now()
	service.Status = "healthy"
	return sr.RegisterService(*service)
}

// UnregisterService 注销服务
func (sr *ServiceRegistry) UnregisterService(serviceName string) error {
	key := fmt.Sprintf("service:%s", serviceName)
	sr.client.Del(sr.ctx, key)
	sr.client.SRem(sr.ctx, "services", serviceName)
	log.Printf("❌ 服务 %s 已注销", serviceName)
	return nil
}

// Close 关闭连接
func (sr *ServiceRegistry) Close() error {
	return sr.client.Close()
}
