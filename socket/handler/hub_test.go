package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

/*
WebSocket Hub功能测试

本文件用于测试Hub结构体的各种功能特性，
包括连接管理、消息广播、事件处理等。

运行命令：
go test -v -run "^Test.*Hub.*$"

测试内容：
1. Hub创建和配置 (NewHub, SetConfig, SetEventHandler等)
2. 连接管理 (AddConnection, RemoveConnection, GetConnection等)
3. 消息处理 (SendMessage, Broadcast, BroadcastWithFilter等)
4. 事件处理 (ConnectionAdded, ConnectionRemoved等)
5. 统计信息 (GetStats, GetConnectionCount等)
6. 错误处理和边界条件
7. 并发访问和性能测试
*/

// TestNewHub 测试 Hub 创建
func TestNewHub(t *testing.T) {
	messageHandler := func(connID string, message []byte) {
		// 测试消息处理器
	}

	hub := NewHub(messageHandler)

	if hub == nil {
		t.Fatal("NewHub returned nil")
	}

	if hub.messageHandler == nil {
		t.Error("Message handler should not be nil")
	}

	if hub.connections == nil {
		t.Error("Connections map should be initialized")
	}

	if hub.broadcastChan == nil {
		t.Error("Broadcast channel should be initialized")
	}

	if hub.ctx == nil {
		t.Error("Context should be initialized")
	}

	if hub.config.MaxConnections != 1000 {
		t.Errorf("Expected MaxConnections to be 1000, got %d", hub.config.MaxConnections)
	}

	if hub.stats == nil {
		t.Error("Stats should be initialized")
	}
}

// TestHubConfiguration 测试 Hub 配置
func TestHubConfiguration(t *testing.T) {
	hub := NewHub(nil)

	// 测试设置事件处理器
	eventHandler := func(event HubEvent, data interface{}) {}
	hub.SetEventHandler(eventHandler)
	if hub.eventHandler == nil {
		t.Error("Event handler should be set")
	}

	// 测试设置配置
	customConfig := HubConfig{
		MaxConnections:    500,
		BroadcastBuffer:   2000,
		CleanupInterval:   2 * time.Minute,
		ConnectionTimeout: 60 * time.Second,
		WriteTimeout:      10 * time.Second,
		MaxMessageSize:    2 * 1024 * 1024,
		MaxConcurrency:    200,
		HeartbeatInterval: 60 * time.Second,
		EnableStats:       false,
	}

	hub.SetConfig(customConfig)
	if hub.config.MaxConnections != 500 {
		t.Errorf("Expected MaxConnections to be 500, got %d", hub.config.MaxConnections)
	}

	if hub.config.BroadcastBuffer != 2000 {
		t.Errorf("Expected BroadcastBuffer to be 2000, got %d", hub.config.BroadcastBuffer)
	}
}

// TestHubLifecycle 测试 Hub 生命周期
func TestHubLifecycle(t *testing.T) {
	hub := NewHub(nil)

	// 测试启动
	err := hub.Start()
	if err != nil {
		t.Fatalf("Failed to start hub: %v", err)
	}

	// 等待一小段时间确保 goroutine 启动
	time.Sleep(10 * time.Millisecond)

	// 测试停止 - 使用 goroutine 和超时
	done := make(chan bool)
	go func() {
		hub.Stop()
		done <- true
	}()

	select {
	case <-done:
		// 停止成功
	case <-time.After(5 * time.Second):
		t.Error("Hub Stop() timed out")
	}

	// 验证上下文被取消
	select {
	case <-hub.ctx.Done():
		// 上下文已取消，这是预期的
	default:
		t.Error("Hub context should be cancelled after Stop()")
	}
}

// TestAddConnection 测试添加连接
func TestAddConnection(t *testing.T) {
	hub := NewHub(nil)
	err := hub.Start()
	if err != nil {
		t.Fatalf("Failed to start hub: %v", err)
	}
	defer hub.Stop()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// 只保持连接，不读取消息
		<-r.Context().Done()
	}))
	defer server.Close()

	// 连接到 WebSocket 服务器
	wsURL := "ws" + server.URL[4:] // 将 http:// 替换为 ws://
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket server: %v", err)
	}
	defer conn.Close()

	connID := "test-conn-1"
	metadata := map[string]interface{}{
		"user_id": "user123",
		"role":    "admin",
	}

	hubConn, err := hub.AddConnection(connID, conn, metadata)
	if err != nil {
		t.Fatalf("Failed to add connection: %v", err)
	}

	if hubConn == nil {
		t.Fatal("Connection should not be nil")
	}

	if hubConn.ID != connID {
		t.Errorf("Expected connection ID to be %s, got %s", connID, hubConn.ID)
	}

	if hubConn.Conn != conn {
		t.Error("Connection WebSocket should match the provided connection")
	}

	if hubConn.Metadata["user_id"] != "user123" {
		t.Error("Connection metadata should be preserved")
	}

	// 验证连接被添加到 Hub
	if hub.GetConnectionCount() != 1 {
		t.Errorf("Expected connection count to be 1, got %d", hub.GetConnectionCount())
	}

	// 验证连接存在
	if _, exists := hub.GetConnection(connID); !exists {
		t.Error("Connection should exist in hub")
	}
}

// TestAddConnectionLimit 测试连接数限制
func TestAddConnectionLimit(t *testing.T) {
	config := DefaultHubConfig()
	config.MaxConnections = 2
	hub := NewHub(nil).SetConfig(config)

	err := hub.Start()
	if err != nil {
		t.Fatalf("Failed to start hub: %v", err)
	}
	defer hub.Stop()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// 只保持连接，不读取消息
		<-r.Context().Done()
	}))
	defer server.Close()

	// 添加第一个连接
	wsURL := "ws" + server.URL[4:] // 将 http:// 替换为 ws://
	conn1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket server: %v", err)
	}
	defer conn1.Close()
	_, err = hub.AddConnection("conn1", conn1, nil)
	if err != nil {
		t.Fatalf("Failed to add first connection: %v", err)
	}

	// 添加第二个连接
	wsURL2 := "ws" + server.URL[4:] // 将 http:// 替换为 ws://
	conn2, _, err := websocket.DefaultDialer.Dial(wsURL2, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket server: %v", err)
	}
	defer conn2.Close()
	_, err = hub.AddConnection("conn2", conn2, nil)
	if err != nil {
		t.Fatalf("Failed to add second connection: %v", err)
	}

	// 尝试添加第三个连接（应该失败）
	conn3, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket server: %v", err)
	}
	defer conn3.Close()
	_, err = hub.AddConnection("conn3", conn3, nil)
	if err == nil {
		t.Error("Expected error when adding connection beyond limit")
	}

	if hub.GetConnectionCount() != 2 {
		t.Errorf("Expected connection count to be 2, got %d", hub.GetConnectionCount())
	}
}

// TestRemoveConnection 测试移除连接
func TestRemoveConnection(t *testing.T) {
	hub := NewHub(nil)
	err := hub.Start()
	if err != nil {
		t.Fatalf("Failed to start hub: %v", err)
	}
	defer hub.Stop()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// 只保持连接，不读取消息
		<-r.Context().Done()
	}))
	defer server.Close()

	wsURL := "ws" + server.URL[4:] // 将 http:// 替换为 ws://
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket server: %v", err)
	}
	defer conn.Close()

	connID := "test-conn-1"
	_, err = hub.AddConnection(connID, conn, nil)
	if err != nil {
		t.Fatalf("Failed to add connection: %v", err)
	}

	// 验证连接存在
	if hub.GetConnectionCount() != 1 {
		t.Errorf("Expected connection count to be 1, got %d", hub.GetConnectionCount())
	}

	// 移除连接
	err = hub.RemoveConnection(connID)
	if err != nil {
		t.Fatalf("Failed to remove connection: %v", err)
	}

	// 验证连接被移除
	if hub.GetConnectionCount() != 0 {
		t.Errorf("Expected connection count to be 0, got %d", hub.GetConnectionCount())
	}

	// 验证连接不存在
	if _, exists := hub.GetConnection(connID); exists {
		t.Error("Connection should not exist after removal")
	}
}

// TestRemoveNonExistentConnection 测试移除不存在的连接
func TestRemoveNonExistentConnection(t *testing.T) {
	hub := NewHub(nil)
	err := hub.Start()
	if err != nil {
		t.Fatalf("Failed to start hub: %v", err)
	}
	defer hub.Stop()

	err = hub.RemoveConnection("non-existent")
	if err == nil {
		t.Error("Expected error when removing non-existent connection")
	}
}

// TestSendMessage 测试发送消息
func TestSendMessage(t *testing.T) {
	hub := NewHub(nil)
	err := hub.Start()
	if err != nil {
		t.Fatalf("Failed to start hub: %v", err)
	}
	defer hub.Stop()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// 只保持连接，不读取消息
		<-r.Context().Done()
	}))
	defer server.Close()

	wsURL := "ws" + server.URL[4:]
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket server: %v", err)
	}
	defer conn.Close()

	connID := "test-conn-1"
	_, err = hub.AddConnection(connID, conn, nil)
	if err != nil {
		t.Fatalf("Failed to add connection: %v", err)
	}

	message := []byte("Hello, World!")
	err = hub.SendMessage(connID, message)
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	// 等待消息处理
	time.Sleep(100 * time.Millisecond)

	// 验证消息被发送（通过统计信息）
	stats := hub.GetStats()
	if stats.TotalMessagesSent != 1 {
		t.Errorf("Expected TotalMessagesSent to be 1, got %d", stats.TotalMessagesSent)
	}
}

// TestSendMessageToNonExistentConnection 测试向不存在的连接发送消息
func TestSendMessageToNonExistentConnection(t *testing.T) {
	hub := NewHub(nil)
	err := hub.Start()
	if err != nil {
		t.Fatalf("Failed to start hub: %v", err)
	}
	defer hub.Stop()

	err = hub.SendMessage("non-existent", []byte("test"))
	if err == nil {
		t.Error("Expected error when sending message to non-existent connection")
	}
}

// TestBroadcast 测试广播功能
func TestBroadcast(t *testing.T) {
	hub := NewHub(nil)
	err := hub.Start()
	if err != nil {
		t.Fatalf("Failed to start hub: %v", err)
	}
	defer hub.Stop()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// 只保持连接，不读取消息
		<-r.Context().Done()
	}))
	defer server.Close()

	// 添加多个连接
	wsURL := "ws" + server.URL[4:] // 将 http:// 替换为 ws://
	conn1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket server: %v", err)
	}
	defer conn1.Close()
	_, err = hub.AddConnection("conn1", conn1, nil)
	if err != nil {
		t.Fatalf("Failed to add first connection: %v", err)
	}

	wsURL2 := "ws" + server.URL[4:] // 将 http:// 替换为 ws://
	conn2, _, err := websocket.DefaultDialer.Dial(wsURL2, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket server: %v", err)
	}
	defer conn2.Close()
	_, err = hub.AddConnection("conn2", conn2, nil)
	if err != nil {
		t.Fatalf("Failed to add second connection: %v", err)
	}

	message := []byte("Broadcast message")
	hub.Broadcast(message)

	// 等待广播处理
	time.Sleep(100 * time.Millisecond)

	// 验证广播统计信息
	stats := hub.GetStats()
	if stats.BroadcastMessages != 1 {
		t.Errorf("Expected BroadcastMessages to be 1, got %d", stats.BroadcastMessages)
	}
}

// TestBroadcastWithFilter 测试带过滤器的广播
func TestBroadcastWithFilter(t *testing.T) {
	hub := NewHub(nil)
	err := hub.Start()
	if err != nil {
		t.Fatalf("Failed to start hub: %v", err)
	}
	defer hub.Stop()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// 只保持连接，不读取消息
		<-r.Context().Done()
	}))
	defer server.Close()

	// 添加连接，设置不同的元数据
	conn1, _, err := websocket.DefaultDialer.Dial(server.URL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket server: %v", err)
	}
	defer conn1.Close()
	_, err = hub.AddConnection("conn1", conn1, map[string]interface{}{"role": "admin"})
	if err != nil {
		t.Fatalf("Failed to add first connection: %v", err)
	}

	wsURL2 := "ws" + server.URL[4:] // 将 http:// 替换为 ws://
	conn2, _, err := websocket.DefaultDialer.Dial(wsURL2, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket server: %v", err)
	}
	defer conn2.Close()
	_, err = hub.AddConnection("conn2", conn2, map[string]interface{}{"role": "user"})
	if err != nil {
		t.Fatalf("Failed to add second connection: %v", err)
	}

	message := []byte("Admin only message")

	// 只向管理员广播
	filter := func(conn *Connection) bool {
		return conn.Metadata["role"] == "admin"
	}

	hub.BroadcastWithFilter(message, filter, nil)

	// 等待广播处理
	time.Sleep(50 * time.Millisecond)

	// 验证只有管理员连接收到消息
	_, msg1, err := conn1.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read message from conn1: %v", err)
	}

	if string(msg1) != string(message) {
		t.Error("Broadcast message content should match")
	}

	// conn2 不应该收到消息
	conn2.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	_, _, err = conn2.ReadMessage()
	if err == nil {
		t.Error("User connection should not receive message")
	}
}

// TestBroadcastWithExclude 测试带排除列表的广播
func TestBroadcastWithExclude(t *testing.T) {
	hub := NewHub(nil)
	err := hub.Start()
	if err != nil {
		t.Fatalf("Failed to start hub: %v", err)
	}
	defer hub.Stop()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// 只保持连接，不读取消息
		<-r.Context().Done()
	}))
	defer server.Close()

	// 添加多个连接
	wsURL := "ws" + server.URL[4:] // 将 http:// 替换为 ws://
	conn1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket server: %v", err)
	}
	defer conn1.Close()
	_, err = hub.AddConnection("conn1", conn1, nil)
	if err != nil {
		t.Fatalf("Failed to add first connection: %v", err)
	}

	wsURL2 := "ws" + server.URL[4:] // 将 http:// 替换为 ws://
	conn2, _, err := websocket.DefaultDialer.Dial(wsURL2, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket server: %v", err)
	}
	defer conn2.Close()
	_, err = hub.AddConnection("conn2", conn2, nil)
	if err != nil {
		t.Fatalf("Failed to add second connection: %v", err)
	}

	message := []byte("Broadcast excluding conn1")

	// 排除 conn1
	hub.BroadcastWithFilter(message, nil, []string{"conn1"})

	// 等待广播处理
	time.Sleep(50 * time.Millisecond)

	// conn1 不应该收到消息
	conn1.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	_, _, err = conn1.ReadMessage()
	if err == nil {
		t.Error("Excluded connection should not receive message")
	}

	// 验证只有 conn2 收到消息
	_, msg2, err := conn2.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read message from conn2: %v", err)
	}

	if string(msg2) != string(message) {
		t.Error("Broadcast message content should match")
	}
}

// TestMessageHandler 测试消息处理器
func TestMessageHandler(t *testing.T) {
	var receivedMessages []string
	var receivedConnIDs []string
	var mu sync.Mutex

	messageHandler := func(connID string, message []byte) {
		mu.Lock()
		defer mu.Unlock()
		receivedConnIDs = append(receivedConnIDs, connID)
		receivedMessages = append(receivedMessages, string(message))
	}

	hub := NewHub(messageHandler)
	err := hub.Start()
	if err != nil {
		t.Fatalf("Failed to start hub: %v", err)
	}
	defer hub.Stop()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// 只保持连接，不读取消息
		<-r.Context().Done()
	}))
	defer server.Close()

	wsURL := "ws" + server.URL[4:] // 将 http:// 替换为 ws://
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket server: %v", err)
	}
	defer conn.Close()

	connID := "test-conn-1"
	_, err = hub.AddConnection(connID, conn, nil)
	if err != nil {
		t.Fatalf("Failed to add connection: %v", err)
	}

	// 发送消息到连接
	testMessage := "Test message"
	err = conn.WriteMessage(websocket.TextMessage, []byte(testMessage))
	if err != nil {
		t.Fatalf("Failed to write message: %v", err)
	}

	// 等待消息处理
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	if len(receivedMessages) != 1 {
		t.Errorf("Expected 1 received message, got %d", len(receivedMessages))
	}

	if len(receivedConnIDs) != 1 {
		t.Errorf("Expected 1 received conn ID, got %d", len(receivedConnIDs))
	}

	if receivedMessages[0] != testMessage {
		t.Errorf("Expected message %s, got %s", testMessage, receivedMessages[0])
	}

	if receivedConnIDs[0] != connID {
		t.Errorf("Expected conn ID %s, got %s", connID, receivedConnIDs[0])
	}
	mu.Unlock()
}

// TestEventHandler 测试事件处理器
func TestEventHandler(t *testing.T) {
	var receivedEvents []HubEvent
	var receivedData []interface{}
	var mu sync.Mutex

	eventHandler := func(event HubEvent, data interface{}) {
		mu.Lock()
		defer mu.Unlock()
		receivedEvents = append(receivedEvents, event)
		receivedData = append(receivedData, data)
	}

	hub := NewHub(nil).SetEventHandler(eventHandler)
	err := hub.Start()
	if err != nil {
		t.Fatalf("Failed to start hub: %v", err)
	}
	defer hub.Stop()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// 只保持连接，不读取消息
		<-r.Context().Done()
	}))
	defer server.Close()

	// 添加连接应该触发事件
	wsURL := "ws" + server.URL[4:] // 将 http:// 替换为 ws://
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket server: %v", err)
	}
	defer conn.Close()
	hubConn, err := hub.AddConnection("test-conn", conn, nil)
	if err != nil {
		t.Fatalf("Failed to add connection: %v", err)
	}

	// 等待事件处理
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	if len(receivedEvents) < 2 { // HubStarted + ConnectionAdded
		t.Errorf("Expected at least 2 events, got %d", len(receivedEvents))
	}

	// 检查 Hub 启动事件
	if receivedEvents[0] != EventHubStarted {
		t.Errorf("Expected first event to be EventHubStarted, got %v", receivedEvents[0])
	}

	// 检查连接添加事件
	if receivedEvents[1] != EventConnectionAdded {
		t.Errorf("Expected second event to be EventConnectionAdded, got %v", receivedEvents[1])
	}

	// 检查连接添加事件的数据
	if receivedData[1] != hubConn {
		t.Error("Connection added event data should be the connection")
	}
	mu.Unlock()

	// 移除连接应该触发事件
	err = hub.RemoveConnection("test-conn")
	if err != nil {
		t.Fatalf("Failed to remove connection: %v", err)
	}

	// 等待事件处理
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	if len(receivedEvents) < 3 {
		t.Errorf("Expected at least 3 events, got %d", len(receivedEvents))
	}

	// 检查连接移除事件
	if receivedEvents[2] != EventConnectionRemoved {
		t.Errorf("Expected third event to be EventConnectionRemoved, got %v", receivedEvents[2])
	}
	mu.Unlock()
}

// TestStats 测试统计信息
func TestStats(t *testing.T) {
	hub := NewHub(nil)
	err := hub.Start()
	if err != nil {
		t.Fatalf("Failed to start hub: %v", err)
	}
	defer hub.Stop()

	stats := hub.GetStats()
	if stats == nil {
		t.Fatal("Stats should not be nil")
	}

	if stats.TotalConnections != 0 {
		t.Errorf("Expected TotalConnections to be 0, got %d", stats.TotalConnections)
	}

	if stats.ActiveConnections != 0 {
		t.Errorf("Expected ActiveConnections to be 0, got %d", stats.ActiveConnections)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// 只保持连接，不读取消息
		<-r.Context().Done()
	}))
	defer server.Close()

	// 添加连接
	wsURL := "ws" + server.URL[4:] // 将 http:// 替换为 ws://
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket server: %v", err)
	}
	defer conn.Close()
	_, err = hub.AddConnection("test-conn", conn, nil)
	if err != nil {
		t.Fatalf("Failed to add connection: %v", err)
	}

	stats = hub.GetStats()
	if stats.TotalConnections != 1 {
		t.Errorf("Expected TotalConnections to be 1, got %d", stats.TotalConnections)
	}

	if stats.ActiveConnections != 1 {
		t.Errorf("Expected ActiveConnections to be 1, got %d", stats.ActiveConnections)
	}

	// 发送消息
	err = hub.SendMessage("test-conn", []byte("test message"))
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	stats = hub.GetStats()
	if stats.TotalMessagesSent != 1 {
		t.Errorf("Expected TotalMessagesSent to be 1, got %d", stats.TotalMessagesSent)
	}

	// 广播消息
	hub.Broadcast([]byte("broadcast message"))
	time.Sleep(50 * time.Millisecond)

	stats = hub.GetStats()
	if stats.BroadcastMessages != 1 {
		t.Errorf("Expected BroadcastMessages to be 1, got %d", stats.BroadcastMessages)
	}
}

// TestConcurrentAccess 测试并发访问
func TestConcurrentAccess(t *testing.T) {
	hub := NewHub(nil)
	err := hub.Start()
	if err != nil {
		t.Fatalf("Failed to start hub: %v", err)
	}
	defer hub.Stop()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// 只保持连接，不读取消息
		<-r.Context().Done()
	}))
	defer server.Close()

	var wg sync.WaitGroup
	numGoroutines := 5
	numConnections := 2

	// 并发添加连接
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numConnections; j++ {
				wsURL := "ws" + server.URL[4:] // 将 http:// 替换为 ws://
				conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
				if err != nil {
					t.Errorf("Failed to connect to WebSocket server: %v", err)
					return
				}
				defer conn.Close()
				connID := fmt.Sprintf("conn-%d-%d", id, j)
				_, err = hub.AddConnection(connID, conn, nil)
				if err != nil {
					t.Errorf("Failed to add connection %s: %v", connID, err)
				}
			}
		}(i)
	}

	wg.Wait()

	expectedConnections := numGoroutines * numConnections
	if hub.GetConnectionCount() != expectedConnections {
		t.Errorf("Expected %d connections, got %d", expectedConnections, hub.GetConnectionCount())
	}
}

// TestConnectionInfo 测试连接信息
func TestConnectionInfo(t *testing.T) {
	hub := NewHub(nil)
	err := hub.Start()
	if err != nil {
		t.Fatalf("Failed to start hub: %v", err)
	}
	defer hub.Stop()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// 只保持连接，不读取消息
		<-r.Context().Done()
	}))
	defer server.Close()

	metadata := map[string]interface{}{
		"user_id": "user123",
		"role":    "admin",
	}

	wsURL := "ws" + server.URL[4:] // 将 http:// 替换为 ws://
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket server: %v", err)
	}
	defer conn.Close()
	connID := "test-conn"
	_, err = hub.AddConnection(connID, conn, metadata)
	if err != nil {
		t.Fatalf("Failed to add connection: %v", err)
	}

	// 获取单个连接信息
	info, err := hub.GetConnectionInfo(connID)
	if err != nil {
		t.Fatalf("Failed to get connection info: %v", err)
	}

	if info.ID != connID {
		t.Errorf("Expected connection ID %s, got %s", connID, info.ID)
	}

	if !info.Connected {
		t.Error("Connection should be marked as connected")
	}

	if info.Metadata["user_id"] != "user123" {
		t.Error("Connection metadata should be preserved")
	}

	// 获取所有连接信息
	allInfos := hub.GetAllConnectionInfo()
	if len(allInfos) != 1 {
		t.Errorf("Expected 1 connection info, got %d", len(allInfos))
	}

	if allInfos[0].ID != connID {
		t.Errorf("Expected connection ID %s, got %s", connID, allInfos[0].ID)
	}
}

// TestHubInfo 测试 Hub 信息
func TestHubInfo(t *testing.T) {
	hub := NewHub(nil)
	err := hub.Start()
	if err != nil {
		t.Fatalf("Failed to start hub: %v", err)
	}
	defer hub.Stop()

	hubInfo := hub.GetHubInfo()
	if hubInfo == nil {
		t.Fatal("Hub info should not be nil")
	}

	if hubInfo.Config.MaxConnections != 1000 {
		t.Errorf("Expected MaxConnections to be 1000, got %d", hubInfo.Config.MaxConnections)
	}

	if hubInfo.Stats == nil {
		t.Error("Stats should not be nil")
	}

	if len(hubInfo.Connections) != 0 {
		t.Errorf("Expected 0 connections, got %d", len(hubInfo.Connections))
	}

	if hubInfo.Uptime == "" {
		t.Error("Uptime should not be empty")
	}
}

// TestMessageSizeLimit 测试消息大小限制
func TestMessageSizeLimit(t *testing.T) {
	config := DefaultHubConfig()
	config.MaxMessageSize = 10 // 设置很小的消息大小限制
	hub := NewHub(nil).SetConfig(config)

	err := hub.Start()
	if err != nil {
		t.Fatalf("Failed to start hub: %v", err)
	}
	defer hub.Stop()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// 只保持连接，不读取消息
		<-r.Context().Done()
	}))
	defer server.Close()

	wsURL := "ws" + server.URL[4:] // 将 http:// 替换为 ws://
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket server: %v", err)
	}
	defer conn.Close()
	_, err = hub.AddConnection("test-conn", conn, nil)
	if err != nil {
		t.Fatalf("Failed to add connection: %v", err)
	}

	// 发送超过限制的消息
	largeMessage := make([]byte, 20) // 超过 10 字节限制
	err = conn.WriteMessage(websocket.TextMessage, largeMessage)
	if err != nil {
		t.Fatalf("Failed to write large message: %v", err)
	}

	// 等待消息处理
	time.Sleep(50 * time.Millisecond)

	// 消息应该被忽略，不会发送到 writeChan
	conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	_, _, err = conn.ReadMessage()
	if err == nil {
		t.Error("Large message should not be processed")
	}
}

// BenchmarkAddConnection 添加连接的性能测试
func BenchmarkAddConnection(b *testing.B) {
	hub := NewHub(nil)
	hub.Start()
	defer hub.Stop()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// 只保持连接，不读取消息
		<-r.Context().Done()
	}))
	defer server.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wsURL := "ws" + server.URL[4:] // 将 http:// 替换为 ws://
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			b.Fatalf("Failed to connect to WebSocket server: %v", err)
		}
		defer conn.Close()
		connID := fmt.Sprintf("conn-%d", i)
		hub.AddConnection(connID, conn, nil)
	}
}

// BenchmarkSendMessage 发送消息的性能测试
func BenchmarkSendMessage(b *testing.B) {
	hub := NewHub(nil)
	hub.Start()
	defer hub.Stop()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// 只保持连接，不读取消息
		<-r.Context().Done()
	}))
	defer server.Close()

	wsURL := "ws" + server.URL[4:] // 将 http:// 替换为 ws://
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		b.Fatalf("Failed to connect to WebSocket server: %v", err)
	}
	defer conn.Close()
	hub.AddConnection("test-conn", conn, nil)

	message := []byte("test message")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hub.SendMessage("test-conn", message)
	}
}

// TestHeartbeatTimeout 测试心跳超时机制
func TestHeartbeatTimeout(t *testing.T) {
	config := DefaultHubConfig()
	config.HeartbeatInterval = 100 * time.Millisecond
	config.ConnectionTimeout = 200 * time.Millisecond
	hub := NewHub(nil).SetConfig(config)

	err := hub.Start()
	if err != nil {
		t.Fatalf("Failed to start hub: %v", err)
	}
	defer hub.Stop()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// 保持连接但不发送心跳
		for {
			select {
			case <-r.Context().Done():
				return
			default:
				time.Sleep(50 * time.Millisecond)
			}
		}
	}))
	defer server.Close()

	wsURL := "ws" + server.URL[4:]
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket server: %v", err)
	}
	defer conn.Close()

	connID := "test-conn"
	_, err = hub.AddConnection(connID, conn, nil)
	if err != nil {
		t.Fatalf("Failed to add connection: %v", err)
	}

	// 验证连接存在
	if hub.GetConnectionCount() != 1 {
		t.Errorf("Expected connection count to be 1, got %d", hub.GetConnectionCount())
	}

	// 等待心跳超时
	time.Sleep(300 * time.Millisecond)

	// 验证连接被移除
	if hub.GetConnectionCount() != 0 {
		t.Errorf("Expected connection count to be 0 after timeout, got %d", hub.GetConnectionCount())
	}
}

// TestCleanupMechanism 测试清理机制
func TestCleanupMechanism(t *testing.T) {
	config := DefaultHubConfig()
	config.CleanupInterval = 100 * time.Millisecond
	config.ConnectionTimeout = 150 * time.Millisecond
	hub := NewHub(nil).SetConfig(config)

	err := hub.Start()
	if err != nil {
		t.Fatalf("Failed to start hub: %v", err)
	}
	defer hub.Stop()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// 保持连接但不发送心跳
		for {
			select {
			case <-r.Context().Done():
				return
			default:
				time.Sleep(50 * time.Millisecond)
			}
		}
	}))
	defer server.Close()

	wsURL := "ws" + server.URL[4:]
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket server: %v", err)
	}
	defer conn.Close()

	connID := "test-conn"
	_, err = hub.AddConnection(connID, conn, nil)
	if err != nil {
		t.Fatalf("Failed to add connection: %v", err)
	}

	// 验证连接存在
	if hub.GetConnectionCount() != 1 {
		t.Errorf("Expected connection count to be 1, got %d", hub.GetConnectionCount())
	}

	// 等待清理机制触发
	time.Sleep(200 * time.Millisecond)

	// 验证连接被清理
	if hub.GetConnectionCount() != 0 {
		t.Errorf("Expected connection count to be 0 after cleanup, got %d", hub.GetConnectionCount())
	}
}

// TestConcurrencyLimit 测试并发限制
func TestConcurrencyLimit(t *testing.T) {
	config := DefaultHubConfig()
	config.MaxConcurrency = 2
	hub := NewHub(nil).SetConfig(config)

	err := hub.Start()
	if err != nil {
		t.Fatalf("Failed to start hub: %v", err)
	}
	defer hub.Stop()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// 只保持连接，不读取消息
		<-r.Context().Done()
	}))
	defer server.Close()

	wsURL := "ws" + server.URL[4:]

	// 添加多个连接
	for i := 0; i < 5; i++ {
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("Failed to connect to WebSocket server: %v", err)
		}
		defer conn.Close()

		connID := fmt.Sprintf("conn-%d", i)
		_, err = hub.AddConnection(connID, conn, nil)
		if err != nil {
			t.Fatalf("Failed to add connection: %v", err)
		}
	}

	// 验证所有连接都被添加
	if hub.GetConnectionCount() != 5 {
		t.Errorf("Expected connection count to be 5, got %d", hub.GetConnectionCount())
	}

	// 广播消息测试并发限制
	message := []byte("test message")
	hub.Broadcast(message)

	// 等待广播处理
	time.Sleep(100 * time.Millisecond)

	// 验证消息被发送（这里主要测试并发限制不会导致死锁）
	stats := hub.GetStats()
	if stats.BroadcastMessages != 1 {
		t.Errorf("Expected BroadcastMessages to be 1, got %d", stats.BroadcastMessages)
	}
}

// TestMessageSizeLimitDetailed 详细测试消息大小限制
func TestMessageSizeLimitDetailed(t *testing.T) {
	config := DefaultHubConfig()
	config.MaxMessageSize = 50 // 设置较小的消息大小限制
	hub := NewHub(nil).SetConfig(config)

	err := hub.Start()
	if err != nil {
		t.Fatalf("Failed to start hub: %v", err)
	}
	defer hub.Stop()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// 只保持连接，不读取消息
		<-r.Context().Done()
	}))
	defer server.Close()

	wsURL := "ws" + server.URL[4:]
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket server: %v", err)
	}
	defer conn.Close()

	connID := "test-conn"
	_, err = hub.AddConnection(connID, conn, nil)
	if err != nil {
		t.Fatalf("Failed to add connection: %v", err)
	}

	// 测试正常大小的消息
	normalMessage := make([]byte, 30) // 小于限制
	err = conn.WriteMessage(websocket.TextMessage, normalMessage)
	if err != nil {
		t.Fatalf("Failed to write normal message: %v", err)
	}

	// 等待消息处理
	time.Sleep(50 * time.Millisecond)

	// 测试超过限制的消息
	largeMessage := make([]byte, 100) // 超过限制
	err = conn.WriteMessage(websocket.TextMessage, largeMessage)
	if err != nil {
		t.Fatalf("Failed to write large message: %v", err)
	}

	// 等待消息处理
	time.Sleep(50 * time.Millisecond)

	// 验证统计信息
	stats := hub.GetStats()
	if stats.TotalMessagesReceived != 1 {
		t.Errorf("Expected TotalMessagesReceived to be 1 (only normal message), got %d", stats.TotalMessagesReceived)
	}
}

// TestConnectionMetadataDetailed 详细测试连接元数据
func TestConnectionMetadataDetailed(t *testing.T) {
	hub := NewHub(nil)
	err := hub.Start()
	if err != nil {
		t.Fatalf("Failed to start hub: %v", err)
	}
	defer hub.Stop()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// 只保持连接，不读取消息
		<-r.Context().Done()
	}))
	defer server.Close()

	wsURL := "ws" + server.URL[4:]
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket server: %v", err)
	}
	defer conn.Close()

	// 测试复杂的元数据
	metadata := map[string]interface{}{
		"user_id":     "user123",
		"role":        "admin",
		"permissions": []string{"read", "write", "delete"},
		"settings": map[string]interface{}{
			"theme":         "dark",
			"language":      "zh-CN",
			"notifications": true,
		},
		"last_login": time.Now(),
		"score":      95.5,
	}

	connID := "test-conn"
	hubConn, err := hub.AddConnection(connID, conn, metadata)
	if err != nil {
		t.Fatalf("Failed to add connection: %v", err)
	}

	// 验证元数据被正确保存
	if hubConn.Metadata["user_id"] != "user123" {
		t.Error("User ID metadata not preserved")
	}

	if hubConn.Metadata["role"] != "admin" {
		t.Error("Role metadata not preserved")
	}

	permissions, ok := hubConn.Metadata["permissions"].([]string)
	if !ok || len(permissions) != 3 {
		t.Error("Permissions metadata not preserved correctly")
	}

	settings, ok := hubConn.Metadata["settings"].(map[string]interface{})
	if !ok || settings["theme"] != "dark" {
		t.Error("Settings metadata not preserved correctly")
	}

	// 测试获取连接信息
	info, err := hub.GetConnectionInfo(connID)
	if err != nil {
		t.Fatalf("Failed to get connection info: %v", err)
	}

	if info.Metadata["user_id"] != "user123" {
		t.Error("Connection info metadata not preserved")
	}
}

// TestErrorRecovery 测试错误恢复机制
func TestErrorRecovery(t *testing.T) {
	hub := NewHub(nil)
	err := hub.Start()
	if err != nil {
		t.Fatalf("Failed to start hub: %v", err)
	}
	defer hub.Stop()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// 模拟连接错误
		conn.Close()
	}))
	defer server.Close()

	wsURL := "ws" + server.URL[4:]
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket server: %v", err)
	}
	defer conn.Close()

	connID := "test-conn"
	_, err = hub.AddConnection(connID, conn, nil)
	if err != nil {
		t.Fatalf("Failed to add connection: %v", err)
	}

	// 等待连接错误处理
	time.Sleep(100 * time.Millisecond)

	// 验证连接被移除
	if hub.GetConnectionCount() != 0 {
		t.Errorf("Expected connection count to be 0 after error, got %d", hub.GetConnectionCount())
	}

	// 测试向已断开的连接发送消息
	err = hub.SendMessage(connID, []byte("test"))
	if err == nil {
		t.Error("Expected error when sending message to disconnected connection")
	}
}

// TestPerformanceStress 性能压力测试
func TestPerformanceStress(t *testing.T) {
	config := DefaultHubConfig()
	config.MaxConnections = 100
	hub := NewHub(nil).SetConfig(config)

	err := hub.Start()
	if err != nil {
		t.Fatalf("Failed to start hub: %v", err)
	}
	defer hub.Stop()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// 只保持连接，不读取消息
		<-r.Context().Done()
	}))
	defer server.Close()

	wsURL := "ws" + server.URL[4:]

	// 并发添加大量连接
	var wg sync.WaitGroup
	numConnections := 50
	for i := 0; i < numConnections; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err != nil {
				t.Errorf("Failed to connect to WebSocket server: %v", err)
				return
			}
			defer conn.Close()

			connID := fmt.Sprintf("conn-%d", id)
			_, err = hub.AddConnection(connID, conn, nil)
			if err != nil {
				t.Errorf("Failed to add connection: %v", err)
			}
		}(i)
	}

	wg.Wait()

	// 验证连接数量
	if hub.GetConnectionCount() != numConnections {
		t.Errorf("Expected connection count to be %d, got %d", numConnections, hub.GetConnectionCount())
	}

	// 并发发送消息
	message := []byte("stress test message")
	for i := 0; i < numConnections; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			connID := fmt.Sprintf("conn-%d", id)
			err := hub.SendMessage(connID, message)
			if err != nil {
				t.Errorf("Failed to send message: %v", err)
			}
		}(i)
	}

	wg.Wait()

	// 验证统计信息
	stats := hub.GetStats()
	if stats.TotalMessagesSent != int64(numConnections) {
		t.Errorf("Expected TotalMessagesSent to be %d, got %d", numConnections, stats.TotalMessagesSent)
	}

	// 并发广播测试
	hub.Broadcast([]byte("broadcast stress test"))
	time.Sleep(100 * time.Millisecond)

	stats = hub.GetStats()
	if stats.BroadcastMessages != 1 {
		t.Errorf("Expected BroadcastMessages to be 1, got %d", stats.BroadcastMessages)
	}
}

// TestEdgeCases 边界条件测试
func TestEdgeCases(t *testing.T) {
	hub := NewHub(nil)
	err := hub.Start()
	if err != nil {
		t.Fatalf("Failed to start hub: %v", err)
	}
	defer hub.Stop()

	// 测试空消息
	err = hub.SendMessage("non-existent", []byte(""))
	if err == nil {
		t.Error("Expected error when sending message to non-existent connection")
	}

	// 测试空广播
	hub.Broadcast([]byte(""))

	// 测试空过滤器
	hub.BroadcastWithFilter([]byte("test"), nil, nil)

	// 测试空排除列表
	hub.BroadcastWithFilter([]byte("test"), nil, []string{})

	// 测试重复添加相同ID的连接
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// 只保持连接，不读取消息
		<-r.Context().Done()
	}))
	defer server.Close()

	wsURL := "ws" + server.URL[4:]
	conn1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket server: %v", err)
	}
	defer conn1.Close()

	conn2, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket server: %v", err)
	}
	defer conn2.Close()

	connID := "duplicate-conn"
	_, err = hub.AddConnection(connID, conn1, nil)
	if err != nil {
		t.Fatalf("Failed to add first connection: %v", err)
	}

	// 尝试添加相同ID的连接
	_, err = hub.AddConnection(connID, conn2, nil)
	if err == nil {
		t.Error("Expected error when adding connection with duplicate ID")
	}

	// 验证只有一个连接
	if hub.GetConnectionCount() != 1 {
		t.Errorf("Expected connection count to be 1, got %d", hub.GetConnectionCount())
	}
}

// TestFilterFunctions 测试过滤器函数
func TestFilterFunctions(t *testing.T) {
	hub := NewHub(nil)
	err := hub.Start()
	if err != nil {
		t.Fatalf("Failed to start hub: %v", err)
	}
	defer hub.Stop()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// 只保持连接，不读取消息
		<-r.Context().Done()
	}))
	defer server.Close()

	wsURL := "ws" + server.URL[4:]

	// 添加不同角色的连接
	roles := []string{"admin", "user", "guest", "moderator"}
	for i, role := range roles {
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("Failed to connect to WebSocket server: %v", err)
		}
		defer conn.Close()

		connID := fmt.Sprintf("conn-%s", role)
		metadata := map[string]interface{}{"role": role, "level": i}
		_, err = hub.AddConnection(connID, conn, metadata)
		if err != nil {
			t.Fatalf("Failed to add connection: %v", err)
		}
	}

	// 测试角色过滤器
	adminFilter := func(conn *Connection) bool {
		return conn.Metadata["role"] == "admin"
	}

	message := []byte("admin only message")
	hub.BroadcastWithFilter(message, adminFilter, nil)

	// 等待广播处理
	time.Sleep(50 * time.Millisecond)

	// 测试级别过滤器
	levelFilter := func(conn *Connection) bool {
		level, ok := conn.Metadata["level"].(int)
		return ok && level >= 2
	}

	message2 := []byte("high level message")
	hub.BroadcastWithFilter(message2, levelFilter, nil)

	// 等待广播处理
	time.Sleep(50 * time.Millisecond)

	// 验证统计信息
	stats := hub.GetStats()
	if stats.BroadcastMessages != 2 {
		t.Errorf("Expected BroadcastMessages to be 2, got %d", stats.BroadcastMessages)
	}
}
