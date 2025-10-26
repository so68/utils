package handler

import (
	"testing"
	"time"
)

func TestNewHub(t *testing.T) {
	messageHandler := func(connID string, message []byte) {
		t.Logf("收到消息: %s", string(message))
	}

	hub := NewHub(messageHandler)
	if hub == nil {
		t.Fatal("Hub 创建失败")
	}

	if hub.connections == nil {
		t.Fatal("连接映射未初始化")
	}

	if hub.messageHandler == nil {
		t.Fatal("消息处理器未设置")
	}
}

func TestHubStartStop(t *testing.T) {
	messageHandler := func(connID string, message []byte) {
		t.Logf("收到消息: %s", string(message))
	}

	hub := NewHub(messageHandler)

	// 启动 Hub
	if err := hub.Start(); err != nil {
		t.Fatalf("启动 Hub 失败: %v", err)
	}

	// 等待一段时间
	time.Sleep(100 * time.Millisecond)

	// 停止 Hub
	hub.Stop()
}

func TestHubConfig(t *testing.T) {
	messageHandler := func(connID string, message []byte) {}
	hub := NewHub(messageHandler)

	config := HubConfig{
		MaxConnections:    5,
		BroadcastBuffer:   100,
		CleanupInterval:   1 * time.Minute,
		ConnectionTimeout: 10 * time.Second,
		EnableStats:       true,
	}

	hub.SetConfig(config)

	if hub.config.MaxConnections != 5 {
		t.Errorf("期望最大连接数为 5，实际为 %d", hub.config.MaxConnections)
	}
}

func TestHubStats(t *testing.T) {
	messageHandler := func(connID string, message []byte) {}
	hub := NewHub(messageHandler)

	// 启动 Hub
	if err := hub.Start(); err != nil {
		t.Fatalf("启动 Hub 失败: %v", err)
	}
	defer hub.Stop()

	// 获取统计信息
	stats := hub.GetStats()
	if stats == nil {
		t.Fatal("统计信息为空")
	}

	if stats.TotalConnections != 0 {
		t.Errorf("期望总连接数为 0，实际为 %d", stats.TotalConnections)
	}
}

func TestConnectionManager(t *testing.T) {
	messageHandler := func(connID string, message []byte) {}
	hub := NewHub(messageHandler)

	connManager := NewConnectionManager(hub)
	if connManager == nil {
		t.Fatal("连接管理器创建失败")
	}

	if connManager.hub != hub {
		t.Fatal("Hub 引用不正确")
	}
}

func TestMessageFilter(t *testing.T) {
	filter := NewMessageFilter()
	if filter == nil {
		t.Fatal("消息过滤器创建失败")
	}

	// 测试允许类型
	filter.AllowType("ping")
	filter.AllowType("pong")

	// 测试阻止类型
	filter.BlockType("spam")

	// 测试过滤
	pingMessage := []byte(`{"type":"ping","data":"test"}`)
	if !filter.Filter("conn1", pingMessage) {
		t.Error("ping 消息应该被允许")
	}

	spamMessage := []byte(`{"type":"spam","data":"test"}`)
	if filter.Filter("conn1", spamMessage) {
		t.Error("spam 消息应该被阻止")
	}
}

func TestRateLimiter(t *testing.T) {
	rateLimiter := NewRateLimiter()
	if rateLimiter == nil {
		t.Fatal("速率限制器创建失败")
	}

	// 设置限制
	rateLimiter.SetLimit("conn1", 100*time.Millisecond)

	// 第一次应该允许
	if !rateLimiter.Allow("conn1") {
		t.Error("第一次应该允许")
	}

	// 立即检查应该被限制
	if rateLimiter.Allow("conn1") {
		t.Error("应该被速率限制")
	}

	// 等待限制期过后
	time.Sleep(150 * time.Millisecond)
	if !rateLimiter.Allow("conn1") {
		t.Error("限制期过后应该允许")
	}

	// 移除限制
	rateLimiter.RemoveLimit("conn1")
	if !rateLimiter.Allow("conn1") {
		t.Error("移除限制后应该允许")
	}
}

func TestJSONMessageHandler(t *testing.T) {
	handler := NewJSONMessageHandler()
	if handler == nil {
		t.Fatal("JSON 消息处理器创建失败")
	}

	// 注册处理器
	pingReceived := false
	handler.RegisterHandler("ping", func(connID string, data map[string]interface{}) {
		pingReceived = true
	})

	// 处理消息
	pingMessage := []byte(`{"type":"ping","data":"test"}`)
	handler.Handle("conn1", pingMessage)

	if !pingReceived {
		t.Error("ping 处理器未被调用")
	}
}

func TestMessageRouter(t *testing.T) {
	router := NewMessageRouter()
	if router == nil {
		t.Fatal("消息路由器创建失败")
	}

	// 添加路由
	routeCalled := false
	router.AddRoute("test", func(connID string, message []byte) error {
		routeCalled = true
		return nil
	})

	// 路由消息
	testMessage := []byte("This is a test message")
	err := router.Route("conn1", testMessage)
	if err != nil {
		t.Errorf("路由失败: %v", err)
	}

	if !routeCalled {
		t.Error("路由处理器未被调用")
	}
}

func TestMessageLogger(t *testing.T) {
	// 使用简单的日志记录器
	logger := &SimpleLogger{}
	messageLogger := NewMessageLogger(logger)
	if messageLogger == nil {
		t.Fatal("消息日志记录器创建失败")
	}

	// 测试日志记录
	messageLogger.LogMessage("conn1", []byte("test message"), "IN")
	messageLogger.LogConnection("conn1", "connected")
	messageLogger.LogError("conn1", nil)
}

// SimpleLogger 简单日志记录器实现
type SimpleLogger struct{}

func (s *SimpleLogger) Debugf(format string, args ...interface{}) {}
func (s *SimpleLogger) Infof(format string, args ...interface{})  {}
func (s *SimpleLogger) Warnf(format string, args ...interface{})  {}
func (s *SimpleLogger) Errorf(format string, args ...interface{}) {}
