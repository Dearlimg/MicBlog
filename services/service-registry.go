package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hashicorp/consul/api"
)

// ServiceRegistry 服务注册中心
type ServiceRegistry struct {
	client  *api.Client
	service *api.AgentServiceRegistration
}

// NewServiceRegistry 创建服务注册中心
func NewServiceRegistry(serviceName, address string, port int) (*ServiceRegistry, error) {
	// 连接Consul
	config := api.DefaultConfig()
	config.Address = os.Getenv("CONSUL_ADDRESS")
	if config.Address == "" {
		config.Address = "localhost:8500"
	}

	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}

	// 注册服务
	service := &api.AgentServiceRegistration{
		ID:      serviceName + "-" + os.Getenv("HOSTNAME"),
		Name:    serviceName,
		Tags:    []string{"microservice", "go"},
		Address: address,
		Port:    port,
		Check: &api.AgentServiceCheck{
			HTTP:     "http://" + address + ":" + string(port) + "/health",
			Interval: "10s",
			Timeout:  "5s",
			TTL:      "",
		},
	}

	err = client.Agent().ServiceRegister(service)
	if err != nil {
		return nil, err
	}

	log.Printf("✅ 服务 %s 已注册到Consul", serviceName)

	return &ServiceRegistry{
		client:  client,
		service: service,
	}, nil
}

// Deregister 注销服务
func (sr *ServiceRegistry) Deregister() error {
	return sr.client.Agent().ServiceDeregister(sr.service.ID)
}

// GetService 获取服务实例列表
func (sr *ServiceRegistry) GetService(serviceName string) ([]*api.ServiceEntry, error) {
	services, _, err := sr.client.Health().Service(serviceName, "", true, nil)
	return services, err
}

// ServiceRegistryExample 服务注册示例
func ServiceRegistryExample() {
	// 注册用户服务
	userRegistry, _ := NewServiceRegistry("user-service", "127.0.0.1", 8001)

	// 注册钱包服务
	walletRegistry, _ := NewServiceRegistry("wallet-service", "127.0.0.1", 8002)

	// 注册评论服务
	commentRegistry, _ := NewServiceRegistry("comment-service", "127.0.0.1", 8003)

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("注销服务...")
	userRegistry.Deregister()
	walletRegistry.Deregister()
	commentRegistry.Deregister()
}

// ServiceDiscoveryExample 服务发现示例
func ServiceDiscoveryExample() {
	config := api.DefaultConfig()
	config.Address = "localhost:8500"

	client, _ := api.NewClient(config)

	// 发现用户服务
	services, _, _ := client.Health().Service("user-service", "", true, nil)

	log.Printf("发现 %d 个用户服务实例:", len(services))
	for _, service := range services {
		log.Printf("  - %s:%d", service.Service.Address, service.Service.Port)
	}
}

// ConsulServiceDiscovery 在API网关中使用服务发现
func ConsulServiceDiscovery() {
	config := api.DefaultConfig()
	config.Address = "localhost:8500"

	client, _ := api.NewClient(config)

	// 动态获取服务地址
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// 发现所有服务
		services, _, _ := client.Health().Service("user-service", "", true, nil)

		log.Printf("发现 %d 个用户服务实例", len(services))
		// 可以在这里更新Nginx配置或直接路由
	}
}
