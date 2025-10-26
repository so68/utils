package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"utils/socket/handler"
)

func main() {
	// 创建消息处理器
	messageHandler := func(connID string, message []byte) {
		fmt.Printf("收到来自 %s 的消息: %s\n", connID, string(message))
	}

	// 创建 Hub
	hub := handler.NewHub(messageHandler)

	// 设置 Hub 配置
	config := handler.HubConfig{
		MaxConnections:    10,
		BroadcastBuffer:   1000,
		CleanupInterval:   2 * time.Minute,
		ConnectionTimeout: 30 * time.Second,
		EnableStats:       true,
	}
	hub.SetConfig(config)

	// 启动 Hub
	if err := hub.Start(); err != nil {
		log.Fatalf("启动 Hub 失败: %v", err)
	}
	defer hub.Stop()

	// 注意：这里需要实际的 WebSocket 连接
	// 在实际使用中，你需要从 HTTP 升级或其他方式获取 websocket.Conn
	log.Println("注意：需要实际的 WebSocket 连接才能测试")
	log.Println("在实际使用中，你需要从 HTTP 升级获取 websocket.Conn")

	// 等待连接建立
	time.Sleep(2 * time.Second)

	// 发送消息到特定连接
	message := []byte("Hello from Hub!")
	if err := hub.SendMessage("conn1", message); err != nil {
		log.Printf("发送消息失败: %v", err)
	}

	// 广播消息
	broadcastMessage := []byte("Broadcast message to all connections!")
	hub.Broadcast(broadcastMessage)

	// 使用连接管理器
	connManager := handler.NewConnectionManager(hub)

	// 向组广播消息
	groupMessage := []byte("Message to group1!")
	if err := connManager.BroadcastToGroup("group1", groupMessage); err != nil {
		log.Printf("组广播失败: %v", err)
	}

	// 设置连接元数据
	if err := connManager.SetConnectionMetadata("conn1", "status", "active"); err != nil {
		log.Printf("设置元数据失败: %v", err)
	}

	// 获取连接信息
	connInfo, err := hub.GetConnectionInfo("conn1")
	if err != nil {
		log.Printf("获取连接信息失败: %v", err)
	} else {
		fmt.Printf("连接信息: %+v\n", connInfo)
	}

	// 获取 Hub 统计信息
	stats := hub.GetStats()
	fmt.Printf("Hub 统计信息: %+v\n", stats)

	// 获取所有连接信息
	allConnections := hub.GetAllConnectionInfo()
	fmt.Printf("所有连接: %d 个\n", len(allConnections))

	// 使用 JSON 消息处理器
	jsonHandler := handler.NewJSONMessageHandler()
	jsonHandler.RegisterHandler("ping", func(connID string, data map[string]interface{}) {
		fmt.Printf("收到 ping 消息来自 %s: %+v\n", connID, data)
	})
	jsonHandler.RegisterHandler("pong", func(connID string, data map[string]interface{}) {
		fmt.Printf("收到 pong 消息来自 %s: %+v\n", connID, data)
	})

	// 发送 JSON 消息
	pingMessage := map[string]interface{}{
		"type":      "ping",
		"data":      "Hello from JSON handler!",
		"timestamp": time.Now().Unix(),
	}
	pingBytes, _ := json.Marshal(pingMessage)
	hub.SendMessage("conn1", pingBytes)

	// 使用消息过滤器
	filter := handler.NewMessageFilter()
	filter.AllowType("ping")
	filter.AllowType("pong")
	filter.BlockType("spam")

	// 使用速率限制器
	rateLimiter := handler.NewRateLimiter()
	rateLimiter.SetLimit("conn1", 1*time.Second)

	// 检查速率限制
	if rateLimiter.Allow("conn1") {
		fmt.Println("允许发送消息")
	} else {
		fmt.Println("速率限制，不允许发送消息")
	}

	// 运行一段时间
	time.Sleep(10 * time.Second)

	// 移除连接
	if err := hub.RemoveConnection("conn1"); err != nil {
		log.Printf("移除连接失败: %v", err)
	}

	fmt.Println("示例完成")
}
