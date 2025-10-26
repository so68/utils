package handler

import (
	"context"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// ConnectionFilter 连接过滤器
type ConnectionFilter func(conn *Connection) bool

// Connection 表示一个 WebSocket 连接
type Connection struct {
	ID       string                 // 连接唯一标识
	Conn     *websocket.Conn        // WebSocket 连接
	Metadata map[string]interface{} // 连接元数据
	Created  time.Time              // 创建时间
	LastSeen time.Time              // 最后活跃时间
	mutex    sync.RWMutex           // 保护元数据的读写锁
	ctx      context.Context        // 连接上下文
	cancel   context.CancelFunc     // 取消函数
}

// BroadcastMessage 广播消息
type BroadcastMessage struct {
	Data    []byte           // 消息数据
	Filter  ConnectionFilter // 连接过滤器，nil表示广播给所有连接
	Exclude []string         // 排除的连接ID列表
}

// HubConfig Hub 配置
type HubConfig struct {
	MaxConnections    int           // 最大连接数，0表示无限制
	BroadcastBuffer   int           // 广播缓冲区大小
	CleanupInterval   time.Duration // 清理间隔
	ConnectionTimeout time.Duration // 连接超时时间
	WriteTimeout      time.Duration // 写超时时间
	MaxMessageSize    int           // 最大消息大小，0表示无限制
	MaxConcurrency    int           // 最大并发数，0表示无限制
	HeartbeatInterval time.Duration // 心跳间隔，0表示不启用
	EnableStats       bool          // 是否启用统计
}

// DefaultHubConfig 返回默认 Hub 配置
func DefaultHubConfig() HubConfig {
	return HubConfig{
		MaxConnections:    1000,
		BroadcastBuffer:   1000,
		CleanupInterval:   5 * time.Minute,
		ConnectionTimeout: 30 * time.Second,
		WriteTimeout:      5 * time.Second,
		MaxMessageSize:    1024 * 1024, // 1MB
		MaxConcurrency:    100,
		HeartbeatInterval: 30 * time.Second,
		EnableStats:       true,
	}
}

// HubStats Hub 统计信息
type HubStats struct {
	TotalConnections      int64     // 总连接数
	ActiveConnections     int64     // 活跃连接数
	TotalMessagesReceived int64     // 总接收消息数
	TotalMessagesSent     int64     // 总发送消息数
	BroadcastMessages     int64     // 广播消息数
	StartTime             time.Time // 启动时间
	LastCleanup           time.Time // 最后清理时间
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
