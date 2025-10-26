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

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源，生产环境中应该检查来源
	},
}

// WebSocketServer WebSocket 服务器
type WebSocketServer struct {
	hub *handler.Hub
}

// NewWebSocketServer 创建 WebSocket 服务器
func NewWebSocketServer() *WebSocketServer {
	// 创建消息处理器
	messageHandler := func(connID string, message []byte) {
		log.Printf("收到来自 %s 的消息: %s", connID, string(message))

		// 简单的回显服务器
		// 注意：这里需要获取 Hub 实例，在实际使用中应该通过参数传递
		log.Printf("处理消息: %s", string(message))
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

	server := &WebSocketServer{
		hub: hub,
	}

	return server
}

// handleWebSocket 处理 WebSocket 连接
func (s *WebSocketServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// 升级 HTTP 连接为 WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket 升级失败: %v", err)
		return
	}
	defer conn.Close()

	// 生成连接ID
	connID := fmt.Sprintf("conn_%d", time.Now().UnixNano())

	// 添加到 Hub
	_, err = s.hub.AddConnection(connID, conn, map[string]interface{}{
		"remote_addr": r.RemoteAddr,
		"user_agent":  r.UserAgent(),
		"created_at":  time.Now(),
	})
	if err != nil {
		log.Printf("添加连接到 Hub 失败: %v", err)
		return
	}

	log.Printf("WebSocket 连接已建立: %s", connID)
}

// handleStats 处理统计信息请求
func (s *WebSocketServer) handleStats(w http.ResponseWriter, r *http.Request) {
	stats := s.hub.GetStats()
	hubInfo := s.hub.GetHubInfo()

	response := map[string]interface{}{
		"stats":       stats,
		"hub_info":    hubInfo,
		"connections": s.hub.GetAllConnectionInfo(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleBroadcast 处理广播消息
func (s *WebSocketServer) handleBroadcast(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "只支持 POST 方法", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Message string `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "无效的 JSON 数据", http.StatusBadRequest)
		return
	}

	// 广播消息
	s.hub.Broadcast([]byte(request.Message))

	response := map[string]interface{}{
		"success": true,
		"message": "消息已广播",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleSendMessage 处理发送消息到特定连接
func (s *WebSocketServer) handleSendMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "只支持 POST 方法", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		ConnID  string `json:"conn_id"`
		Message string `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "无效的 JSON 数据", http.StatusBadRequest)
		return
	}

	// 发送消息到指定连接
	err := s.hub.SendMessage(request.ConnID, []byte(request.Message))
	if err != nil {
		http.Error(w, fmt.Sprintf("发送消息失败: %v", err), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "消息已发送",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleConnections 处理连接列表请求
func (s *WebSocketServer) handleConnections(w http.ResponseWriter, r *http.Request) {
	connections := s.hub.GetAllConnectionInfo()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(connections)
}

// Start 启动服务器
func (s *WebSocketServer) Start() error {
	// 启动 Hub
	if err := s.hub.Start(); err != nil {
		return fmt.Errorf("启动 Hub 失败: %v", err)
	}

	// 设置路由
	http.HandleFunc("/ws", s.handleWebSocket)
	http.HandleFunc("/api/stats", s.handleStats)
	http.HandleFunc("/api/broadcast", s.handleBroadcast)
	http.HandleFunc("/api/send", s.handleSendMessage)
	http.HandleFunc("/api/connections", s.handleConnections)

	// 静态文件服务
	http.Handle("/", http.FileServer(http.Dir("./static/")))

	return nil
}

// Stop 停止服务器
func (s *WebSocketServer) Stop() {
	s.hub.Stop()
}

func mainWebSocketServer() {
	// 创建 WebSocket 服务器
	server := NewWebSocketServer()

	// 启动服务器
	if err := server.Start(); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
	defer server.Stop()

	// 启动 HTTP 服务器
	log.Println("WebSocket 服务器启动在 :8080")
	log.Println("WebSocket 端点: ws://localhost:8080/ws")
	log.Println("统计信息: http://localhost:8080/api/stats")
	log.Println("连接列表: http://localhost:8080/api/connections")
	log.Println("广播消息: POST http://localhost:8080/api/broadcast")
	log.Println("发送消息: POST http://localhost:8080/api/send")

	log.Fatal(http.ListenAndServe(":8080", nil))
}
