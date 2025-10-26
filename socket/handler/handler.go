package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// DefaultMessageHandler 默认消息处理器
func DefaultMessageHandler(connID string, message []byte) {
	log.Printf("Received message from %s: %s", connID, string(message))
}

// JSONMessageHandler JSON 消息处理器
type JSONMessageHandler struct {
	handlers map[string]func(connID string, data map[string]interface{})
}

// NewJSONMessageHandler 创建 JSON 消息处理器
func NewJSONMessageHandler() *JSONMessageHandler {
	return &JSONMessageHandler{
		handlers: make(map[string]func(connID string, data map[string]interface{})),
	}
}

// RegisterHandler 注册消息处理器
func (h *JSONMessageHandler) RegisterHandler(messageType string, handler func(connID string, data map[string]interface{})) {
	h.handlers[messageType] = handler
}

// Handle 处理消息
func (h *JSONMessageHandler) Handle(connID string, message []byte) {
	var data map[string]interface{}
	if err := json.Unmarshal(message, &data); err != nil {
		log.Printf("Failed to parse JSON message from %s: %v", connID, err)
		return
	}

	messageType, ok := data["type"].(string)
	if !ok {
		log.Printf("Message type not found in message from %s", connID)
		return
	}

	handler, exists := h.handlers[messageType]
	if !exists {
		log.Printf("No handler registered for message type: %s", messageType)
		return
	}

	handler(connID, data)
}

// MessageRouterImpl 消息路由器实现
type MessageRouterImpl struct {
	routes map[string]func(connID string, message []byte) error
}

// NewMessageRouter 创建消息路由器
func NewMessageRouter() *MessageRouterImpl {
	return &MessageRouterImpl{
		routes: make(map[string]func(connID string, message []byte) error),
	}
}

// AddRoute 添加路由
func (r *MessageRouterImpl) AddRoute(pattern string, handler func(connID string, message []byte) error) {
	r.routes[pattern] = handler
}

// Route 路由消息
func (r *MessageRouterImpl) Route(connID string, message []byte) error {
	// 简单的字符串匹配路由
	for pattern, handler := range r.routes {
		if strings.Contains(string(message), pattern) {
			return handler(connID, message)
		}
	}

	// 默认处理器
	log.Printf("No route found for message from %s: %s", connID, string(message))
	return nil
}

// ConnectionManager 连接管理器
type ConnectionManager struct {
	hub *Hub
}

// NewConnectionManager 创建连接管理器
func NewConnectionManager(hub *Hub) *ConnectionManager {
	return &ConnectionManager{
		hub: hub,
	}
}

// AddConnectionWithRetry 带重试的连接添加
func (cm *ConnectionManager) AddConnectionWithRetry(connID string, wsConn *websocket.Conn, metadata map[string]interface{}, maxRetries int) error {
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		_, err := cm.hub.AddConnection(connID, wsConn, metadata)
		if err == nil {
			return nil
		}

		lastErr = err
		log.Printf("Connection attempt %d failed for %s: %v", i+1, connID, err)

		if i < maxRetries-1 {
			time.Sleep(time.Duration(i+1) * time.Second) // 递增延迟
		}
	}

	return fmt.Errorf("failed to add connection after %d retries: %v", maxRetries, lastErr)
}

// BroadcastToGroup 向组广播消息
func (cm *ConnectionManager) BroadcastToGroup(groupName string, message []byte) error {
	connections := cm.hub.GetConnections()

	var wg sync.WaitGroup
	successCount := 0

	for _, conn := range connections {
		conn.mutex.RLock()
		group, exists := conn.Metadata["group"]
		conn.mutex.RUnlock()

		if exists && group == groupName {
			wg.Add(1)
			go func(c *Connection) {
				defer wg.Done()
				if c.Conn != nil {
					if err := c.Conn.WriteMessage(websocket.TextMessage, message); err == nil {
						successCount++
					}
				}
			}(conn)
		}
	}

	wg.Wait()
	log.Printf("Broadcasted to group %s: %d successful, %d total", groupName, successCount, len(connections))
	return nil
}

// GetConnectionsByGroup 获取组内连接
func (cm *ConnectionManager) GetConnectionsByGroup(groupName string) []*Connection {
	connections := cm.hub.GetConnections()
	var groupConnections []*Connection

	for _, conn := range connections {
		conn.mutex.RLock()
		group, exists := conn.Metadata["group"]
		conn.mutex.RUnlock()

		if exists && group == groupName {
			groupConnections = append(groupConnections, conn)
		}
	}

	return groupConnections
}

// SetConnectionMetadata 设置连接元数据
func (cm *ConnectionManager) SetConnectionMetadata(connID string, key string, value interface{}) error {
	conn, exists := cm.hub.GetConnection(connID)
	if !exists {
		return fmt.Errorf("connection not found: %s", connID)
	}

	conn.mutex.Lock()
	conn.Metadata[key] = value
	conn.mutex.Unlock()

	return nil
}

// GetConnectionMetadata 获取连接元数据
func (cm *ConnectionManager) GetConnectionMetadata(connID string, key string) (interface{}, error) {
	conn, exists := cm.hub.GetConnection(connID)
	if !exists {
		return nil, fmt.Errorf("connection not found: %s", connID)
	}

	conn.mutex.RLock()
	value, exists := conn.Metadata[key]
	conn.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("metadata key not found: %s", key)
	}

	return value, nil
}

// MessageFilter 消息过滤器
type MessageFilter struct {
	allowedTypes []string
	blockedTypes []string
}

// NewMessageFilter 创建消息过滤器
func NewMessageFilter() *MessageFilter {
	return &MessageFilter{
		allowedTypes: make([]string, 0),
		blockedTypes: make([]string, 0),
	}
}

// AllowType 允许消息类型
func (f *MessageFilter) AllowType(messageType string) {
	f.allowedTypes = append(f.allowedTypes, messageType)
}

// BlockType 阻止消息类型
func (f *MessageFilter) BlockType(messageType string) {
	f.blockedTypes = append(f.blockedTypes, messageType)
}

// Filter 过滤消息
func (f *MessageFilter) Filter(connID string, message []byte) bool {
	var data map[string]interface{}
	if err := json.Unmarshal(message, &data); err != nil {
		return false // 无法解析的消息被过滤
	}

	messageType, ok := data["type"].(string)
	if !ok {
		return false // 没有类型的消息被过滤
	}

	// 检查阻止列表
	for _, blockedType := range f.blockedTypes {
		if messageType == blockedType {
			return false
		}
	}

	// 检查允许列表
	if len(f.allowedTypes) > 0 {
		for _, allowedType := range f.allowedTypes {
			if messageType == allowedType {
				return true
			}
		}
		return false
	}

	return true
}

// RateLimiter 速率限制器
type RateLimiter struct {
	limits map[string]*time.Ticker
	mutex  sync.RWMutex
}

// NewRateLimiter 创建速率限制器
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		limits: make(map[string]*time.Ticker),
	}
}

// SetLimit 设置限制
func (rl *RateLimiter) SetLimit(connID string, interval time.Duration) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	// 停止旧的限制器
	if ticker, exists := rl.limits[connID]; exists {
		ticker.Stop()
	}

	// 创建新的限制器
	rl.limits[connID] = time.NewTicker(interval)
}

// Allow 检查是否允许发送
func (rl *RateLimiter) Allow(connID string) bool {
	rl.mutex.RLock()
	ticker, exists := rl.limits[connID]
	rl.mutex.RUnlock()

	if !exists {
		return true // 没有限制
	}

	select {
	case <-ticker.C:
		return true
	default:
		return false
	}
}

// RemoveLimit 移除限制
func (rl *RateLimiter) RemoveLimit(connID string) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	if ticker, exists := rl.limits[connID]; exists {
		ticker.Stop()
		delete(rl.limits, connID)
	}
}

// MessageLogger 消息日志记录器
type MessageLogger struct {
	logger Logger
}

// Logger 日志接口
type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

// NewMessageLogger 创建消息日志记录器
func NewMessageLogger(logger Logger) *MessageLogger {
	return &MessageLogger{
		logger: logger,
	}
}

// LogMessage 记录消息
func (ml *MessageLogger) LogMessage(connID string, message []byte, direction string) {
	ml.logger.Infof("[%s] %s message from %s: %s", direction, connID, string(message))
}

// LogConnection 记录连接事件
func (ml *MessageLogger) LogConnection(connID string, event string) {
	ml.logger.Infof("Connection %s: %s", connID, event)
}

// LogError 记录错误
func (ml *MessageLogger) LogError(connID string, err error) {
	ml.logger.Errorf("Error for connection %s: %v", connID, err)
}
