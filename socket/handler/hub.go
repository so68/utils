package handler

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

// HubEvent Hub 事件类型
type HubEvent int

// MessageHandler 消息处理器类型
type MessageHandler func(connID string, message []byte)

// EventHandler 事件处理器类型
type EventHandler func(event HubEvent, data interface{})

const (
	EventConnectionAdded   HubEvent = iota // 连接添加事件
	EventConnectionRemoved                 // 连接移除事件
	EventMessageReceived                   // 消息接收事件
	EventBroadcastSent                     // 广播发送事件
	EventHubStarted                        // Hub 启动事件
	EventHubStopped                        // Hub 停止事件
)

// Hub 管理多个 WebSocket 连接的中心管理器
type Hub struct {
	logger *slog.Logger // 日志记录器

	// 连接管理
	connections map[string]*Connection // 连接映射，key为连接ID
	connMutex   sync.RWMutex           // 保护连接映射的读写锁

	// 消息处理
	messageHandler MessageHandler // 全局消息处理器
	eventHandler   EventHandler   // 事件处理器
	broadcastChan  chan []byte    // 广播消息通道

	// 生命周期管理
	ctx    context.Context    // 上下文
	cancel context.CancelFunc // 取消函数
	wg     sync.WaitGroup     // 等待组

	// 配置
	config HubConfig // Hub 配置

	// 统计信息
	stats *HubStats // 统计信息
}

// NewHub 创建新的 Hub 实例
func NewHub(messageHandler MessageHandler) *Hub {
	ctx, cancel := context.WithCancel(context.Background())

	config := DefaultHubConfig()
	hub := &Hub{
		logger:         slog.Default(),
		connections:    make(map[string]*Connection),
		messageHandler: messageHandler,
		eventHandler:   nil,
		broadcastChan:  make(chan []byte, config.BroadcastBuffer),
		ctx:            ctx,
		cancel:         cancel,
		config:         config,
		stats: &HubStats{
			StartTime: time.Now(),
		},
	}

	return hub
}

// SetEventHandler 设置事件处理器
func (h *Hub) SetEventHandler(eventHandler EventHandler) *Hub {
	h.eventHandler = eventHandler
	return h
}

// SetConfig 设置 Hub 配置
func (h *Hub) SetConfig(config HubConfig) *Hub {
	h.config = config
	return h
}

// SetLogger 设置日志记录器
func (h *Hub) SetLogger(logger *slog.Logger) *Hub {
	h.logger = logger
	return h
}

// Start 启动 Hub
func (h *Hub) Start() error {
	// 启动广播处理器
	h.wg.Add(1)
	go h.broadcastLoop()

	// 启动清理器
	if h.config.CleanupInterval > 0 {
		h.wg.Add(1)
		go h.cleanupLoop()
	}

	// 启动心跳检查器
	if h.config.HeartbeatInterval > 0 {
		h.wg.Add(1)
		go h.heartbeatLoop()
	}

	// 更新统计信息
	if h.config.EnableStats {
		h.stats.StartTime = time.Now()
	}

	// 触发 Hub 启动事件
	if h.eventHandler != nil {
		h.eventHandler(EventHubStarted, nil)
	}

	h.logger.Info("Hub started")
	return nil
}

// Stop 停止 Hub
func (h *Hub) Stop() {
	h.logger.Info("Stopping Hub...")

	// 先取消 Hub 上下文，通知所有 goroutine 退出
	h.cancel()

	// 取消所有连接的上下文，通知连接 goroutine 退出
	h.connMutex.Lock()
	for _, conn := range h.connections {
		if conn.cancel != nil {
			conn.cancel()
		}
	}
	h.connMutex.Unlock()

	// 等待所有 goroutine 退出
	h.wg.Wait()

	// 然后关闭所有 WebSocket 连接
	h.connMutex.Lock()
	for _, conn := range h.connections {
		if conn.Conn != nil {
			conn.Conn.Close()
		}
	}
	h.connections = make(map[string]*Connection)
	h.connMutex.Unlock()

	// 安全关闭广播通道
	select {
	case <-h.broadcastChan:
		// 通道已关闭
	default:
		close(h.broadcastChan)
	}

	// 触发 Hub 停止事件
	if h.eventHandler != nil {
		h.eventHandler(EventHubStopped, nil)
	}

	h.logger.Info("Hub stopped")
}

// AddConnection 添加连接
func (h *Hub) AddConnection(connID string, wsConn *websocket.Conn, metadata map[string]interface{}) (*Connection, error) {
	// 检查连接数限制（使用写锁确保原子性）
	h.connMutex.Lock()
	defer h.connMutex.Unlock()

	if h.config.MaxConnections > 0 && len(h.connections) >= h.config.MaxConnections {
		return nil, fmt.Errorf("maximum connections limit reached: %d", h.config.MaxConnections)
	}

	// 检查连接ID是否已存在
	if _, exists := h.connections[connID]; exists {
		return nil, fmt.Errorf("connection with ID %s already exists", connID)
	}

	// 创建连接对象
	connCtx, connCancel := context.WithCancel(h.ctx)
	conn := &Connection{
		ID:       connID,
		Conn:     wsConn,
		Metadata: metadata,
		Created:  time.Now(),
		LastSeen: time.Now(),
		ctx:      connCtx,
		cancel:   connCancel,
	}

	// 添加到连接映射
	h.connections[connID] = conn

	// 启动消息监听
	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
		h.listenConnection(conn)
	}()

	// 更新统计信息
	if h.config.EnableStats {
		atomic.AddInt64(&h.stats.TotalConnections, 1)
		atomic.AddInt64(&h.stats.ActiveConnections, 1)
	}

	// 触发连接添加事件
	if h.eventHandler != nil {
		h.eventHandler(EventConnectionAdded, conn)
	}

	h.logger.Info("Connection added", "conn_id", connID)
	return conn, nil
}

// RemoveConnection 移除连接
func (h *Hub) RemoveConnection(connID string) error {
	h.connMutex.Lock()
	conn, exists := h.connections[connID]
	if exists {
		delete(h.connections, connID)
	}
	h.connMutex.Unlock()

	if !exists {
		return fmt.Errorf("connection not found: %s", connID)
	}

	// 取消连接的上下文，通知 goroutine 退出
	if conn.cancel != nil {
		conn.cancel()
	}
	// 关闭 WebSocket 连接
	if conn.Conn != nil {
		conn.Conn.Close()
	}

	// 更新统计信息
	if h.config.EnableStats {
		atomic.AddInt64(&h.stats.ActiveConnections, -1)
	}

	// 触发连接移除事件
	if h.eventHandler != nil {
		h.eventHandler(EventConnectionRemoved, conn)
	}

	h.logger.Info("Connection removed", "conn_id", connID)
	return nil
}

// SendMessage 发送消息到指定连接
func (h *Hub) SendMessage(connID string, message []byte) error {
	conn, exists := h.GetConnection(connID)
	if !exists {
		return fmt.Errorf("connection not found: %s", connID)
	}

	if conn.Conn == nil {
		return fmt.Errorf("websocket connection is nil for connection: %s", connID)
	}

	// 更新最后活跃时间
	conn.mutex.Lock()
	conn.LastSeen = time.Now()
	conn.mutex.Unlock()

	// 发送消息
	conn.Conn.SetWriteDeadline(time.Now().Add(h.config.WriteTimeout))
	if err := conn.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
		// 发送失败时移除连接
		h.RemoveConnection(connID)
		return fmt.Errorf("failed to send message to %s: %v", connID, err)
	}
	// 重置写超时
	conn.Conn.SetWriteDeadline(time.Time{})

	// 更新统计信息
	if h.config.EnableStats {
		atomic.AddInt64(&h.stats.TotalMessagesSent, 1)
	}

	return nil
}

// Broadcast 广播消息
func (h *Hub) Broadcast(message []byte) {
	select {
	case h.broadcastChan <- message:
	default:
		h.logger.Warn("Broadcast channel is full, dropping message")
	}
}

// BroadcastWithFilter 带过滤器的广播
func (h *Hub) BroadcastWithFilter(message []byte, filter ConnectionFilter, exclude []string) {
	excludeMap := make(map[string]bool)
	for _, id := range exclude {
		excludeMap[id] = true
	}

	h.connMutex.RLock()
	connections := make([]*Connection, 0, len(h.connections))
	for _, conn := range h.connections {
		if excludeMap[conn.ID] {
			continue
		}
		if filter == nil || filter(conn) {
			connections = append(connections, conn)
		}
	}
	h.connMutex.RUnlock()

	// 并发发送消息
	var wg sync.WaitGroup
	var semaphore chan struct{}
	var failedConnections []string
	var failedMutex sync.Mutex

	// 如果设置了并发限制，创建信号量
	if h.config.MaxConcurrency > 0 {
		semaphore = make(chan struct{}, h.config.MaxConcurrency)
	}

	for _, conn := range connections {
		wg.Add(1)
		go func(c *Connection) {
			defer wg.Done()

			// 如果设置了并发限制，获取信号量
			if semaphore != nil {
				semaphore <- struct{}{}
				defer func() { <-semaphore }()
			}

			if c.Conn != nil {
				// 设置写超时
				c.Conn.SetWriteDeadline(time.Now().Add(h.config.WriteTimeout))
				if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
					h.logger.Warn("Failed to send broadcast message", "conn_id", c.ID, "error", err.Error())
					// 记录失败的连接，稍后统一移除
					failedMutex.Lock()
					failedConnections = append(failedConnections, c.ID)
					failedMutex.Unlock()
				} else {
					// 重置写超时
					c.Conn.SetWriteDeadline(time.Time{})
				}
			}
		}(conn)
	}
	wg.Wait()

	// 统一移除失败的连接
	if len(failedConnections) > 0 {
		for _, connID := range failedConnections {
			h.RemoveConnection(connID)
		}
	}

	// 触发广播发送事件
	if h.eventHandler != nil {
		h.eventHandler(EventBroadcastSent, map[string]interface{}{
			"message":     message,
			"connections": len(connections),
			"filter":      filter != nil,
			"exclude":     exclude,
		})
	}

	// 更新统计信息
	if h.config.EnableStats {
		atomic.AddInt64(&h.stats.BroadcastMessages, 1)
	}
}

// GetConnection 获取连接
func (h *Hub) GetConnection(connID string) (*Connection, bool) {
	h.connMutex.RLock()
	conn, exists := h.connections[connID]
	h.connMutex.RUnlock()
	return conn, exists
}

// GetConnectionCount 获取连接数量
func (h *Hub) GetConnectionCount() int {
	h.connMutex.RLock()
	count := len(h.connections)
	h.connMutex.RUnlock()
	return count
}

// GetConnections 获取所有连接（返回连接ID列表，避免大量内存复制）
func (h *Hub) GetConnections() []string {
	h.connMutex.RLock()
	connIDs := make([]string, 0, len(h.connections))
	for connID := range h.connections {
		connIDs = append(connIDs, connID)
	}
	h.connMutex.RUnlock()
	return connIDs
}

// GetStats 获取统计信息
func (h *Hub) GetStats() *HubStats {
	// 创建统计信息副本
	stats := &HubStats{
		TotalConnections:      atomic.LoadInt64(&h.stats.TotalConnections),
		ActiveConnections:     atomic.LoadInt64(&h.stats.ActiveConnections),
		TotalMessagesReceived: atomic.LoadInt64(&h.stats.TotalMessagesReceived),
		TotalMessagesSent:     atomic.LoadInt64(&h.stats.TotalMessagesSent),
		BroadcastMessages:     atomic.LoadInt64(&h.stats.BroadcastMessages),
		StartTime:             h.stats.StartTime,
		LastCleanup:           h.stats.LastCleanup,
	}

	return stats
}

// GetConnectionInfo 获取连接信息
func (h *Hub) GetConnectionInfo(connID string) (*ConnectionInfo, error) {
	conn, exists := h.GetConnection(connID)
	if !exists {
		return nil, fmt.Errorf("connection not found: %s", connID)
	}

	conn.mutex.RLock()
	metadata := make(map[string]interface{})
	for k, v := range conn.Metadata {
		metadata[k] = v
	}
	conn.mutex.RUnlock()

	info := &ConnectionInfo{
		ID:        conn.ID,
		URL:       "", // WebSocket 连接没有 URL 信息
		Connected: conn.Conn != nil,
		Created:   conn.Created,
		LastSeen:  conn.LastSeen,
		Metadata:  metadata,
		Stats:     make(map[string]interface{}), // 简化统计信息
	}

	return info, nil
}

// GetAllConnectionInfo 获取所有连接信息
func (h *Hub) GetAllConnectionInfo() []ConnectionInfo {
	h.connMutex.RLock()
	connections := make([]*Connection, 0, len(h.connections))
	for _, conn := range h.connections {
		connections = append(connections, conn)
	}
	h.connMutex.RUnlock()

	infos := make([]ConnectionInfo, 0, len(connections))
	for _, conn := range connections {
		conn.mutex.RLock()
		metadata := make(map[string]interface{})
		for k, v := range conn.Metadata {
			metadata[k] = v
		}
		conn.mutex.RUnlock()

		info := ConnectionInfo{
			ID:        conn.ID,
			URL:       "", // WebSocket 连接没有 URL 信息
			Connected: conn.Conn != nil,
			Created:   conn.Created,
			LastSeen:  conn.LastSeen,
			Metadata:  metadata,
			Stats:     make(map[string]interface{}), // 简化统计信息
		}
		infos = append(infos, info)
	}

	return infos
}

// GetHubInfo 获取 Hub 信息
func (h *Hub) GetHubInfo() *HubInfo {
	return &HubInfo{
		Config:      h.config,
		Stats:       h.GetStats(),
		Connections: h.GetAllConnectionInfo(),
		Uptime:      time.Since(h.stats.StartTime).String(),
	}
}

// listenConnection 监听连接消息
func (h *Hub) listenConnection(conn *Connection) {
	defer func() {
		// 连接断开时从 Hub 中移除（使用原子操作避免竞态条件）
		h.connMutex.Lock()
		if _, exists := h.connections[conn.ID]; exists {
			delete(h.connections, conn.ID)
			// 更新统计信息
			if h.config.EnableStats {
				atomic.AddInt64(&h.stats.ActiveConnections, -1)
			}
			h.logger.Info("Connection disconnected", "conn_id", conn.ID)
		}
		h.connMutex.Unlock()
	}()

	for {
		select {
		case <-conn.ctx.Done():
			return
		default:
			// 读取消息
			messageType, message, err := conn.Conn.ReadMessage()
			if err != nil {
				h.logger.Error("Read error for connection", "conn_id", conn.ID, "error", err.Error())
				return
			}

			// 只处理文本消息
			if messageType == websocket.TextMessage {
				h.hubMessageHandler(conn.ID, message)
			}
		}
	}
}

// hubMessageHandler 处理接收到的消息
func (h *Hub) hubMessageHandler(connID string, message []byte) {
	// 检查消息大小限制
	if h.config.MaxMessageSize > 0 && len(message) > h.config.MaxMessageSize {
		h.logger.Warn("Message too large", "conn_id", connID, "size", len(message), "limit", h.config.MaxMessageSize)
		return
	}

	// 更新连接的最后活跃时间
	if conn, exists := h.GetConnection(connID); exists {
		conn.mutex.Lock()
		conn.LastSeen = time.Now()
		conn.mutex.Unlock()
	}

	// 调用消息处理器
	if h.messageHandler != nil {
		h.messageHandler(connID, message)
	}

	// 触发消息接收事件
	if h.eventHandler != nil {
		h.eventHandler(EventMessageReceived, map[string]interface{}{
			"conn_id": connID,
			"message": message,
		})
	}

	// 更新统计信息
	if h.config.EnableStats {
		atomic.AddInt64(&h.stats.TotalMessagesReceived, 1)
	}
}

// broadcastLoop 广播循环
func (h *Hub) broadcastLoop() {
	defer h.wg.Done()

	for {
		select {
		case <-h.ctx.Done():
			return
		case message, ok := <-h.broadcastChan:
			if !ok {
				return
			}
			h.BroadcastWithFilter(message, nil, nil)
		}
	}
}

// cleanupLoop 清理循环
func (h *Hub) cleanupLoop() {
	defer h.wg.Done()

	ticker := time.NewTicker(h.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-h.ctx.Done():
			return
		case <-ticker.C:
			h.cleanup()
		}
	}
}

// heartbeatLoop 心跳检查循环
func (h *Hub) heartbeatLoop() {
	defer h.wg.Done()

	ticker := time.NewTicker(h.config.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-h.ctx.Done():
			return
		case <-ticker.C:
			h.checkHeartbeat()
		}
	}
}

// checkHeartbeat 检查连接心跳
func (h *Hub) checkHeartbeat() {
	h.connMutex.RLock()
	connections := make([]*Connection, 0, len(h.connections))
	for _, conn := range h.connections {
		connections = append(connections, conn)
	}
	h.connMutex.RUnlock()

	now := time.Now()
	timeout := h.config.ConnectionTimeout

	for _, conn := range connections {
		conn.mutex.RLock()
		lastSeen := conn.LastSeen
		conn.mutex.RUnlock()

		// 检查连接是否超时
		if now.Sub(lastSeen) > timeout {
			h.logger.Warn("Connection heartbeat timeout", "conn_id", conn.ID, "last_seen", lastSeen)
			h.RemoveConnection(conn.ID)
		}
	}
}

// cleanup 清理无效连接
func (h *Hub) cleanup() {
	h.connMutex.Lock()
	defer h.connMutex.Unlock()

	now := time.Now()
	toRemove := make([]string, 0)

	for connID, conn := range h.connections {
		// 检查连接是否超时
		if now.Sub(conn.LastSeen) > h.config.ConnectionTimeout {
			toRemove = append(toRemove, connID)
			continue
		}

		// 检查 WebSocket 连接状态
		if conn.Conn != nil {
			// 设置写超时来检查连接状态
			conn.Conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
			if err := conn.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				toRemove = append(toRemove, connID)
				continue
			}
			// 重置写超时
			conn.Conn.SetWriteDeadline(time.Time{})
		} else {
			toRemove = append(toRemove, connID)
		}
	}

	// 移除无效连接
	for _, connID := range toRemove {
		if conn, exists := h.connections[connID]; exists {
			if conn.Conn != nil {
				conn.Conn.Close()
			}
			delete(h.connections, connID)

			// 更新统计信息
			if h.config.EnableStats {
				atomic.AddInt64(&h.stats.ActiveConnections, -1)
			}

			h.logger.Info("Cleaned up connection", "conn_id", connID)
		}
	}

	// 更新清理时间
	if h.config.EnableStats {
		h.stats.LastCleanup = now
	}

	if len(toRemove) > 0 {
		h.logger.Info("Cleanup completed", "removed_connections", len(toRemove))
	}
}
