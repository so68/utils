package handler

import (
	"context"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Hub 管理多个 WebSocket 连接的中心管理器
type Hub struct {
	// 连接管理
	connections map[string]*Connection // 连接映射，key为连接ID
	connMutex   sync.RWMutex           // 保护连接映射的读写锁

	// 消息处理
	messageHandler MessageHandler // 全局消息处理器
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

// Connection 表示一个 WebSocket 连接
type Connection struct {
	ID       string                 // 连接唯一标识
	Conn     *websocket.Conn        // WebSocket 连接
	Metadata map[string]interface{} // 连接元数据
	Created  time.Time              // 创建时间
	LastSeen time.Time              // 最后活跃时间
	mutex    sync.RWMutex           // 保护元数据的读写锁
}

// HubConfig Hub 配置
type HubConfig struct {
	MaxConnections    int           // 最大连接数，0表示无限制
	BroadcastBuffer   int           // 广播缓冲区大小
	CleanupInterval   time.Duration // 清理间隔
	ConnectionTimeout time.Duration // 连接超时时间
	EnableStats       bool          // 是否启用统计
}

// DefaultHubConfig 返回默认 Hub 配置
func DefaultHubConfig() HubConfig {
	return HubConfig{
		MaxConnections:    1000,
		BroadcastBuffer:   1000,
		CleanupInterval:   5 * time.Minute,
		ConnectionTimeout: 30 * time.Second,
		EnableStats:       true,
	}
}

// HubStats Hub 统计信息
type HubStats struct {
	TotalConnections  int64        // 总连接数
	ActiveConnections int64        // 活跃连接数
	TotalMessages     int64        // 总消息数
	BroadcastMessages int64        // 广播消息数
	StartTime         time.Time    // 启动时间
	LastCleanup       time.Time    // 最后清理时间
	mutex             sync.RWMutex // 保护统计信息的读写锁
}

// MessageHandler 消息处理器类型
type MessageHandler func(connID string, message []byte)

// ConnectionHandler 连接处理器类型
type ConnectionHandler func(conn *Connection) error

// HubEvent Hub 事件类型
type HubEvent int

const (
	EventConnectionAdded HubEvent = iota
	EventConnectionRemoved
	EventMessageReceived
	EventBroadcastSent
	EventHubStarted
	EventHubStopped
)

// HubEventHandler Hub 事件处理器
type HubEventHandler func(event HubEvent, data interface{})

// ConnectionFilter 连接过滤器
type ConnectionFilter func(conn *Connection) bool

// MessageRouterInterface 消息路由器接口
type MessageRouterInterface interface {
	Route(connID string, message []byte) error
}

// BroadcastMessage 广播消息
type BroadcastMessage struct {
	Data    []byte           // 消息数据
	Filter  ConnectionFilter // 连接过滤器，nil表示广播给所有连接
	Exclude []string         // 排除的连接ID列表
}

// ConnectionInfo 连接信息
type ConnectionInfo struct {
	ID        string                 `json:"id"`
	URL       string                 `json:"url"`
	Connected bool                   `json:"connected"`
	Created   time.Time              `json:"created"`
	LastSeen  time.Time              `json:"last_seen"`
	Metadata  map[string]interface{} `json:"metadata"`
	Stats     map[string]interface{} `json:"stats"`
}

// HubInfo Hub 信息
type HubInfo struct {
	Config      HubConfig        `json:"config"`
	Stats       *HubStats        `json:"stats"`
	Connections []ConnectionInfo `json:"connections"`
	Uptime      string           `json:"uptime"`
}
