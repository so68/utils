package handler

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

// NewHub 创建新的 Hub 实例
func NewHub(messageHandler MessageHandler) *Hub {
	ctx, cancel := context.WithCancel(context.Background())

	hub := &Hub{
		connections:    make(map[string]*Connection),
		messageHandler: messageHandler,
		broadcastChan:  make(chan []byte, 1000),
		ctx:            ctx,
		cancel:         cancel,
		config:         DefaultHubConfig(),
		stats: &HubStats{
			StartTime: time.Now(),
		},
	}

	return hub
}

// SetConfig 设置 Hub 配置
func (h *Hub) SetConfig(config HubConfig) {
	h.config = config
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

	// 更新统计信息
	if h.config.EnableStats {
		h.stats.StartTime = time.Now()
	}

	log.Printf("Hub started with config: %+v", h.config)
	return nil
}

// Stop 停止 Hub
func (h *Hub) Stop() {
	log.Println("Stopping Hub...")

	// 取消上下文
	h.cancel()

	// 关闭所有连接
	h.connMutex.Lock()
	for _, conn := range h.connections {
		if conn.Conn != nil {
			conn.Conn.Close()
		}
	}
	h.connections = make(map[string]*Connection)
	h.connMutex.Unlock()

	// 关闭广播通道
	close(h.broadcastChan)

	// 等待所有 goroutine 完成
	h.wg.Wait()

	log.Println("Hub stopped")
}

// AddConnection 添加连接
func (h *Hub) AddConnection(connID string, wsConn *websocket.Conn, metadata map[string]interface{}) (*Connection, error) {
	// 检查连接数限制
	if h.config.MaxConnections > 0 {
		h.connMutex.RLock()
		connCount := len(h.connections)
		h.connMutex.RUnlock()

		if connCount >= h.config.MaxConnections {
			return nil, fmt.Errorf("maximum connections limit reached: %d", h.config.MaxConnections)
		}
	}

	// 创建连接对象
	conn := &Connection{
		ID:       connID,
		Conn:     wsConn,
		Metadata: metadata,
		Created:  time.Now(),
		LastSeen: time.Now(),
	}

	// 添加到连接映射
	h.connMutex.Lock()
	h.connections[connID] = conn
	h.connMutex.Unlock()

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

	log.Printf("Connection added: %s", connID)
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

	// 关闭 WebSocket 连接
	if conn.Conn != nil {
		conn.Conn.Close()
	}

	// 更新统计信息
	if h.config.EnableStats {
		atomic.AddInt64(&h.stats.ActiveConnections, -1)
	}

	log.Printf("Connection removed: %s", connID)
	return nil
}

// GetConnection 获取连接
func (h *Hub) GetConnection(connID string) (*Connection, bool) {
	h.connMutex.RLock()
	conn, exists := h.connections[connID]
	h.connMutex.RUnlock()
	return conn, exists
}

// GetConnections 获取所有连接
func (h *Hub) GetConnections() map[string]*Connection {
	h.connMutex.RLock()
	connections := make(map[string]*Connection)
	for k, v := range h.connections {
		connections[k] = v
	}
	h.connMutex.RUnlock()
	return connections
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
	if err := conn.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
		return fmt.Errorf("failed to send message to %s: %v", connID, err)
	}

	// 更新统计信息
	if h.config.EnableStats {
		atomic.AddInt64(&h.stats.TotalMessages, 1)
	}

	return nil
}

// Broadcast 广播消息
func (h *Hub) Broadcast(message []byte) {
	select {
	case h.broadcastChan <- message:
	default:
		log.Printf("Broadcast channel is full, dropping message")
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
	for _, conn := range connections {
		wg.Add(1)
		go func(c *Connection) {
			defer wg.Done()
			if c.Conn != nil {
				c.Conn.WriteMessage(websocket.TextMessage, message)
			}
		}(conn)
	}
	wg.Wait()

	// 更新统计信息
	if h.config.EnableStats {
		atomic.AddInt64(&h.stats.BroadcastMessages, 1)
	}
}

// GetStats 获取统计信息
func (h *Hub) GetStats() *HubStats {
	h.stats.mutex.RLock()
	defer h.stats.mutex.RUnlock()

	// 创建统计信息副本
	stats := &HubStats{
		TotalConnections:  atomic.LoadInt64(&h.stats.TotalConnections),
		ActiveConnections: atomic.LoadInt64(&h.stats.ActiveConnections),
		TotalMessages:     atomic.LoadInt64(&h.stats.TotalMessages),
		BroadcastMessages: atomic.LoadInt64(&h.stats.BroadcastMessages),
		StartTime:         h.stats.StartTime,
		LastCleanup:       h.stats.LastCleanup,
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
		// 连接断开时从 Hub 中移除
		h.connMutex.Lock()
		delete(h.connections, conn.ID)
		h.connMutex.Unlock()

		// 更新统计信息
		if h.config.EnableStats {
			atomic.AddInt64(&h.stats.ActiveConnections, -1)
		}

		log.Printf("Connection %s disconnected", conn.ID)
	}()

	for {
		select {
		case <-h.ctx.Done():
			return
		default:
			// 读取消息
			messageType, message, err := conn.Conn.ReadMessage()
			if err != nil {
				log.Printf("Read error for connection %s: %v", conn.ID, err)
				return
			}

			// 只处理文本消息
			if messageType == websocket.TextMessage {
				// 更新最后活跃时间
				conn.mutex.Lock()
				conn.LastSeen = time.Now()
				conn.mutex.Unlock()

				// 调用消息处理器
				if h.messageHandler != nil {
					h.messageHandler(conn.ID, message)
				}

				// 更新统计信息
				if h.config.EnableStats {
					atomic.AddInt64(&h.stats.TotalMessages, 1)
				}
			}
		}
	}
}

// handleMessage 处理接收到的消息
func (h *Hub) handleMessage(connID string, message []byte) {
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

	// 更新统计信息
	if h.config.EnableStats {
		atomic.AddInt64(&h.stats.TotalMessages, 1)
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
			// 尝试发送 ping 来检查连接状态
			if err := conn.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				toRemove = append(toRemove, connID)
			}
		}
	}

	// 移除无效连接
	for _, connID := range toRemove {
		conn := h.connections[connID]
		if conn.Conn != nil {
			conn.Conn.Close()
		}
		delete(h.connections, connID)

		// 更新统计信息
		if h.config.EnableStats {
			atomic.AddInt64(&h.stats.ActiveConnections, -1)
		}

		log.Printf("Cleaned up connection: %s", connID)
	}

	// 更新清理时间
	if h.config.EnableStats {
		h.stats.mutex.Lock()
		h.stats.LastCleanup = now
		h.stats.mutex.Unlock()
	}

	if len(toRemove) > 0 {
		log.Printf("Cleanup completed, removed %d connections", len(toRemove))
	}
}
