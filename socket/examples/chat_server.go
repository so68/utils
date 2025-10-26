package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"utils/socket/handler"

	"github.com/gorilla/websocket"
)

// ChatMessage 聊天消息结构
type ChatMessage struct {
	Type      string    `json:"type"`
	User      string    `json:"user"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Room      string    `json:"room"`
}

// ChatServer 聊天服务器
type ChatServer struct {
	hub           *handler.Hub
	connManager   *handler.ConnectionManager
	jsonHandler   *handler.JSONMessageHandler
	rateLimiter   *handler.RateLimiter
	messageLogger *handler.MessageLogger
}

// NewChatServer 创建聊天服务器
func NewChatServer() *ChatServer {
	// 创建消息处理器
	messageHandler := func(connID string, message []byte) {
		// 这里可以添加消息处理逻辑
		log.Printf("收到来自 %s 的消息: %s", connID, string(message))
	}

	// 创建 Hub
	hub := handler.NewHub(messageHandler)

	// 设置配置
	config := handler.HubConfig{
		MaxConnections:    100,
		BroadcastBuffer:   1000,
		CleanupInterval:   5 * time.Minute,
		ConnectionTimeout: 30 * time.Second,
		EnableStats:       true,
	}
	hub.SetConfig(config)

	// 创建连接管理器
	connManager := handler.NewConnectionManager(hub)

	// 创建 JSON 消息处理器
	jsonHandler := handler.NewJSONMessageHandler()

	// 创建速率限制器
	rateLimiter := handler.NewRateLimiter()

	// 创建消息日志记录器
	messageLogger := handler.NewMessageLogger(&SimpleLogger{})

	server := &ChatServer{
		hub:           hub,
		connManager:   connManager,
		jsonHandler:   jsonHandler,
		rateLimiter:   rateLimiter,
		messageLogger: messageLogger,
	}

	// 注册消息处理器
	server.registerHandlers()

	return server
}

// registerHandlers 注册消息处理器
func (cs *ChatServer) registerHandlers() {
	// 注册聊天消息处理器
	cs.jsonHandler.RegisterHandler("chat", func(connID string, data map[string]interface{}) {
		cs.handleChatMessage(connID, data)
	})

	// 注册加入房间处理器
	cs.jsonHandler.RegisterHandler("join_room", func(connID string, data map[string]interface{}) {
		cs.handleJoinRoom(connID, data)
	})

	// 注册离开房间处理器
	cs.jsonHandler.RegisterHandler("leave_room", func(connID string, data map[string]interface{}) {
		cs.handleLeaveRoom(connID, data)
	})

	// 注册 ping 处理器
	cs.jsonHandler.RegisterHandler("ping", func(connID string, data map[string]interface{}) {
		cs.handlePing(connID, data)
	})
}

// handleChatMessage 处理聊天消息
func (cs *ChatServer) handleChatMessage(connID string, data map[string]interface{}) {
	// 检查速率限制
	if !cs.rateLimiter.Allow(connID) {
		cs.sendError(connID, "消息发送过于频繁，请稍后再试")
		return
	}

	// 获取用户信息
	user, err := cs.connManager.GetConnectionMetadata(connID, "user")
	if err != nil {
		cs.sendError(connID, "用户信息未找到")
		return
	}

	// 获取房间信息
	room, err := cs.connManager.GetConnectionMetadata(connID, "room")
	if err != nil {
		cs.sendError(connID, "房间信息未找到")
		return
	}

	// 创建聊天消息
	chatMsg := ChatMessage{
		Type:      "chat",
		User:      user.(string),
		Message:   data["message"].(string),
		Timestamp: time.Now(),
		Room:      room.(string),
	}

	// 序列化消息
	msgBytes, err := json.Marshal(chatMsg)
	if err != nil {
		cs.sendError(connID, "消息序列化失败")
		return
	}

	// 向房间内所有用户广播消息
	cs.connManager.BroadcastToGroup(room.(string), msgBytes)

	// 记录消息日志
	cs.messageLogger.LogMessage(connID, msgBytes, "OUT")
}

// handleJoinRoom 处理加入房间
func (cs *ChatServer) handleJoinRoom(connID string, data map[string]interface{}) {
	room, ok := data["room"].(string)
	if !ok {
		cs.sendError(connID, "房间名称无效")
		return
	}

	// 设置房间元数据
	if err := cs.connManager.SetConnectionMetadata(connID, "room", room); err != nil {
		cs.sendError(connID, "设置房间失败")
		return
	}

	// 设置组
	if err := cs.connManager.SetConnectionMetadata(connID, "group", room); err != nil {
		cs.sendError(connID, "设置组失败")
		return
	}

	// 发送加入成功消息
	joinMsg := map[string]interface{}{
		"type":      "join_success",
		"room":      room,
		"timestamp": time.Now(),
	}
	joinBytes, _ := json.Marshal(joinMsg)
	cs.hub.SendMessage(connID, joinBytes)

	// 向房间内其他用户广播新用户加入
	notification := map[string]interface{}{
		"type":      "user_joined",
		"user":      connID,
		"room":      room,
		"timestamp": time.Now(),
	}
	notifBytes, _ := json.Marshal(notification)
	cs.connManager.BroadcastToGroup(room, notifBytes)

	cs.messageLogger.LogConnection(connID, fmt.Sprintf("加入房间: %s", room))
}

// handleLeaveRoom 处理离开房间
func (cs *ChatServer) handleLeaveRoom(connID string, data map[string]interface{}) {
	// 获取当前房间
	room, err := cs.connManager.GetConnectionMetadata(connID, "room")
	if err != nil {
		cs.sendError(connID, "未在房间中")
		return
	}

	// 向房间内其他用户广播用户离开
	notification := map[string]interface{}{
		"type":      "user_left",
		"user":      connID,
		"room":      room,
		"timestamp": time.Now(),
	}
	notifBytes, _ := json.Marshal(notification)
	cs.connManager.BroadcastToGroup(room.(string), notifBytes)

	// 清除房间信息
	cs.connManager.SetConnectionMetadata(connID, "room", "")
	cs.connManager.SetConnectionMetadata(connID, "group", "")

	cs.messageLogger.LogConnection(connID, fmt.Sprintf("离开房间: %s", room))
}

// handlePing 处理 ping 消息
func (cs *ChatServer) handlePing(connID string, data map[string]interface{}) {
	pongMsg := map[string]interface{}{
		"type":      "pong",
		"timestamp": time.Now(),
	}
	pongBytes, _ := json.Marshal(pongMsg)
	cs.hub.SendMessage(connID, pongBytes)
}

// sendError 发送错误消息
func (cs *ChatServer) sendError(connID string, message string) {
	errorMsg := map[string]interface{}{
		"type":      "error",
		"message":   message,
		"timestamp": time.Now(),
	}
	errorBytes, _ := json.Marshal(errorMsg)
	cs.hub.SendMessage(connID, errorBytes)
}

// Start 启动聊天服务器
func (cs *ChatServer) Start() error {
	// 启动 Hub
	if err := cs.hub.Start(); err != nil {
		return fmt.Errorf("启动 Hub 失败: %v", err)
	}

	// 设置速率限制（每秒最多 10 条消息）
	cs.rateLimiter.SetLimit("default", 100*time.Millisecond)

	return nil
}

// Stop 停止聊天服务器
func (cs *ChatServer) Stop() {
	cs.hub.Stop()
}

// AddUser 添加用户连接
func (cs *ChatServer) AddUser(connID string, wsConn *websocket.Conn, username string) error {
	// 添加连接
	_, err := cs.hub.AddConnection(connID, wsConn, map[string]interface{}{
		"user":  username,
		"group": "",
	})
	if err != nil {
		return err
	}

	// 设置用户元数据
	cs.connManager.SetConnectionMetadata(connID, "user", username)

	// 设置速率限制
	cs.rateLimiter.SetLimit(connID, 100*time.Millisecond)

	cs.messageLogger.LogConnection(connID, fmt.Sprintf("用户 %s 连接", username))
	return nil
}

// RemoveUser 移除用户连接
func (cs *ChatServer) RemoveUser(connID string) error {
	// 获取用户信息
	user, _ := cs.connManager.GetConnectionMetadata(connID, "user")

	// 移除连接
	err := cs.hub.RemoveConnection(connID)
	if err != nil {
		return err
	}

	// 移除速率限制
	cs.rateLimiter.RemoveLimit(connID)

	cs.messageLogger.LogConnection(connID, fmt.Sprintf("用户 %s 断开连接", user))
	return nil
}

// GetStats 获取服务器统计信息
func (cs *ChatServer) GetStats() *handler.HubStats {
	return cs.hub.GetStats()
}

// GetRoomUsers 获取房间用户列表
func (cs *ChatServer) GetRoomUsers(room string) []string {
	connections := cs.connManager.GetConnectionsByGroup(room)
	users := make([]string, 0, len(connections))

	for _, conn := range connections {
		if user, err := cs.connManager.GetConnectionMetadata(conn.ID, "user"); err == nil {
			users = append(users, user.(string))
		}
	}

	return users
}

// HTTP 处理函数
func (cs *ChatServer) handleStats(w http.ResponseWriter, r *http.Request) {
	stats := cs.GetStats()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (cs *ChatServer) handleRooms(w http.ResponseWriter, r *http.Request) {
	// 这里可以实现获取房间列表的逻辑
	rooms := []string{"general", "tech", "random"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rooms)
}

func mainChatServer() {
	// 创建聊天服务器
	chatServer := NewChatServer()

	// 启动服务器
	if err := chatServer.Start(); err != nil {
		log.Fatalf("启动聊天服务器失败: %v", err)
	}
	defer chatServer.Stop()

	// 设置 HTTP 路由
	http.HandleFunc("/stats", chatServer.handleStats)
	http.HandleFunc("/rooms", chatServer.handleRooms)

	// 启动 HTTP 服务器
	go func() {
		log.Println("HTTP 服务器启动在 :8080")
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	// 模拟添加一些用户连接
	// 注意：这里需要实际的 WebSocket 服务器来测试
	// 在实际使用中，你需要连接到真实的 WebSocket 服务器

	log.Println("聊天服务器已启动")
	log.Println("访问 http://localhost:8080/stats 查看统计信息")
	log.Println("访问 http://localhost:8080/rooms 查看房间列表")

	// 保持程序运行
	select {}
}

// SimpleLogger 简单日志记录器实现
type SimpleLogger struct{}

func (s *SimpleLogger) Debugf(format string, args ...interface{}) {
	log.Printf("[DEBUG] "+format, args...)
}
func (s *SimpleLogger) Infof(format string, args ...interface{}) {
	log.Printf("[INFO] "+format, args...)
}
func (s *SimpleLogger) Warnf(format string, args ...interface{}) {
	log.Printf("[WARN] "+format, args...)
}
func (s *SimpleLogger) Errorf(format string, args ...interface{}) {
	log.Printf("[ERROR] "+format, args...)
}
