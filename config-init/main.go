package main

import (
	"blog/shared/config"
	"log"
)

func main() {
	// 创建配置中心
	configCenter, err := config.NewConfigCenter("47.118.19.28:6379", "sta_go", 0)
	if err != nil {
		log.Fatalf("Failed to create config center: %v", err)
	}
	defer configCenter.Close()

	// 初始化默认配置
	err = configCenter.InitializeDefaultConfigs()
	if err != nil {
		log.Fatalf("Failed to initialize default configs: %v", err)
	}

	log.Println("Default configurations initialized successfully!")
}
