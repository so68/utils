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

// 优化后的Hub结构
type OptimizedHub struct {
	// 连接管理
	connections map[string]*Connection
	connMutex   sync.RWMutex

	// 消息处理
	messageHandler MessageHandler
	broadcastChan  chan []byte

	// 生命周期管理
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// 配置
	config HubConfig

	// 统计信息
	stats *HubStats

	// 优化：连接池和清理优化
	cleanupTicker *time.Ticker
	cleanupDone   chan struct{}

	// 优化：批量操作
	batchSize    int
	cleanupBatch chan []string
}

// 创建优化后的Hub
func NewOptimizedHub(messageHandler MessageHandler) *OptimizedHub {
	ctx, cancel := context.WithCancel(context.Background())

	hub := &OptimizedHub{
		connections:    make(map[string]*Connection),
		messageHandler: messageHandler,
		broadcastChan:  make(chan []byte, 1000),
		ctx:            ctx,
		cancel:         cancel,
		config:         DefaultHubConfig(),
		stats: &HubStats{
			StartTime: time.Now(),
		},
		cleanupDone:  make(chan struct{}),
		batchSize:    100, // 批量处理大小
		cleanupBatch: make(chan []string, 10),
	}

	return hub
}

// 优化的清理方法
func (h *OptimizedHub) optimizedCleanup() {
	h.connMutex.Lock()
	defer h.connMutex.Unlock()

	now := time.Now()
	toRemove := make([]string, 0, h.batchSize)

	// 使用更高效的清理策略
	for connID, conn := range h.connections {
		// 检查连接是否超时
		if now.Sub(conn.LastSeen) > h.config.ConnectionTimeout {
			toRemove = append(toRemove, connID)
			continue
		}

		// 优化：异步检查连接状态，避免阻塞
		if conn.Conn != nil {
			// 使用非阻塞的ping检查
			select {
			case <-time.After(100 * time.Millisecond): // 100ms超时
				// 超时说明连接可能有问题
				toRemove = append(toRemove, connID)
			default:
				// 尝试发送ping，但不等待响应
				go func(connID string, wsConn *websocket.Conn) {
					if err := wsConn.WriteMessage(websocket.PingMessage, nil); err != nil {
						// 异步移除连接
						select {
						case h.cleanupBatch <- []string{connID}:
						default:
							// 如果通道满了，直接记录日志
							log.Printf("Failed to queue connection %s for cleanup", connID)
						}
					}
				}(connID, conn.Conn)
			}
		}

		// 如果批量大小达到，处理一批
		if len(toRemove) >= h.batchSize {
			h.processCleanupBatch(toRemove)
			toRemove = toRemove[:0] // 重置切片
		}
	}

	// 处理剩余的连接
	if len(toRemove) > 0 {
		h.processCleanupBatch(toRemove)
	}

	// 更新清理时间
	if h.config.EnableStats {
		h.stats.mutex.Lock()
		h.stats.LastCleanup = now
		h.stats.mutex.Unlock()
	}
}

// 批量处理清理
func (h *OptimizedHub) processCleanupBatch(toRemove []string) {
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
}

// 优化的广播方法
func (h *OptimizedHub) OptimizedBroadcast(message []byte) {
	select {
	case h.broadcastChan <- message:
	default:
		// 如果通道满了，记录警告但不阻塞
		log.Printf("Broadcast channel is full, dropping message")
	}
}

// 优化的带过滤器的广播
func (h *OptimizedHub) OptimizedBroadcastWithFilter(message []byte, filter ConnectionFilter, exclude []string) {
	// 构建排除映射
	excludeMap := make(map[string]bool, len(exclude))
	for _, id := range exclude {
		excludeMap[id] = true
	}

	// 获取连接列表
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

	// 使用工作池模式并发发送
	workerCount := 10 // 可配置的工作协程数
	if len(connections) < workerCount {
		workerCount = len(connections)
	}

	if workerCount == 0 {
		return
	}

	// 创建任务通道
	taskChan := make(chan *Connection, len(connections))
	var wg sync.WaitGroup

	// 启动工作协程
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for conn := range taskChan {
				if conn.Conn != nil {
					// 使用超时发送，避免阻塞
					done := make(chan error, 1)
					go func() {
						done <- conn.Conn.WriteMessage(websocket.TextMessage, message)
					}()

					select {
					case <-done:
						// 发送完成
					case <-time.After(5 * time.Second):
						// 发送超时
						log.Printf("Message send timeout for connection %s", conn.ID)
					}
				}
			}
		}()
	}

	// 分发任务
	for _, conn := range connections {
		taskChan <- conn
	}
	close(taskChan)

	// 等待所有工作协程完成
	wg.Wait()

	// 更新统计信息
	if h.config.EnableStats {
		atomic.AddInt64(&h.stats.BroadcastMessages, 1)
	}
}

// 优化的连接添加方法
func (h *OptimizedHub) OptimizedAddConnection(connID string, wsConn *websocket.Conn, metadata map[string]interface{}) (*Connection, error) {
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
		h.optimizedListenConnection(conn)
	}()

	// 更新统计信息
	if h.config.EnableStats {
		atomic.AddInt64(&h.stats.TotalConnections, 1)
		atomic.AddInt64(&h.stats.ActiveConnections, 1)
	}

	log.Printf("Connection added: %s", connID)
	return conn, nil
}

// 优化的连接监听
func (h *OptimizedHub) optimizedListenConnection(conn *Connection) {
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

	// 设置读取超时
	conn.Conn.SetReadDeadline(time.Now().Add(h.config.ConnectionTimeout))

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

				// 重置读取超时
				conn.Conn.SetReadDeadline(time.Now().Add(h.config.ConnectionTimeout))

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
