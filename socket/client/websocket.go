package client

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

// Websocket 通用 WebSocket 管理器
type Websocket struct {
	conn              *websocket.Conn         // 连接
	dialer            *websocket.Dialer       // 拨号器
	config            Config                  // 配置
	messageHandler    MessageHandler          // 消息处理器
	beforeConnHandler BeforeConnectionHandler // 连接前的回调处理器
	afterConnHandler  AfterConnectionHandler  // 连接成功后的回调处理器
	logger            Logger                  // 日志记录器
	metrics           Metrics                 // 性能指标
	ctx               context.Context         // 上下文
	cancel            context.CancelFunc      // 取消函数
	isRunning         bool                    // 是否运行
	retryCount        int                     // 重试次数
	dialURL           string                  // 保存原始连接URL用于重连
	mux               sync.RWMutex            // 保护并发访问
	messageCount      int64                   // 消息计数器
	startTime         time.Time               // 启动时间
	goroutines        sync.WaitGroup          // 管理goroutine生命周期
}

// NewWebsocket 创建WebSocket实例
func NewWebsocket(dialURL string, messageHandler MessageHandler) *Websocket {
	dialer := websocket.DefaultDialer
	ctx, cancel := context.WithCancel(context.Background())

	m := &Websocket{
		dialer:         dialer,
		config:         DefaultConfig(),
		messageHandler: messageHandler,
		logger:         &NoopLogger{},
		metrics:        &NoopMetrics{},
		ctx:            ctx,
		cancel:         cancel,
		dialURL:        dialURL,
		startTime:      time.Now(),
	}
	return m
}

// Start 运行WebSocket
func (m *Websocket) Start() error {
	if err := m.connect(m.dialURL); err != nil {
		return err
	}

	// 启动心跳和监听goroutine
	m.goroutines.Add(2)
	go func() {
		defer m.goroutines.Done()
		m.pingLoop()
	}()
	go func() {
		defer m.goroutines.Done()
		m.listenLoop()
	}()

	// 安全地设置运行状态
	m.mux.Lock()
	m.isRunning = true
	m.mux.Unlock()

	return nil
}

// SetConfig 设置配置
func (m *Websocket) SetConfig(config Config) {
	m.config = config
}

// SetLogger 设置日志记录器
func (m *Websocket) SetLogger(logger Logger) {
	m.logger = logger
}

// SetMetrics 设置性能指标
func (m *Websocket) SetMetrics(metrics Metrics) {
	m.metrics = metrics
}

// SetBeforeConnectionHandler 设置连接前的回调处理器
func (m *Websocket) SetBeforeConnectionHandler(handler BeforeConnectionHandler) {
	m.beforeConnHandler = handler
}

// SetAfterConnectionHandler 设置连接成功后的回调处理器
func (m *Websocket) SetAfterConnectionHandler(handler AfterConnectionHandler) {
	m.afterConnHandler = handler
}

// WriteMessage 发送消息
func (m *Websocket) WriteMessage(message []byte) error {
	m.mux.Lock()
	defer m.mux.Unlock()
	if m.conn == nil {
		return fmt.Errorf("connection is not established")
	}
	if err := m.conn.WriteMessage(websocket.TextMessage, message); err != nil {
		return fmt.Errorf("write message failed: %v", err)
	}
	return nil
}

// connect 连接或重连
func (m *Websocket) connect(dialURL string) error {
	// 执行连接前的回调
	if m.beforeConnHandler != nil {
		if err := m.beforeConnHandler(m); err != nil {
			m.logger.Errorf("Before connection handler error: %v", err)
			return fmt.Errorf("before connection handler failed: %v", err)
		}
		m.logger.Debugf("Before connection handler executed successfully")
	}

	reqHeader := http.Header{}

	// 添加自定义请求头
	for key, value := range m.config.Headers {
		reqHeader.Add(key, value)
	}

	// 设置连接超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 使用带超时的拨号器
	dialer := &websocket.Dialer{
		HandshakeTimeout: 30 * time.Second,
		ReadBufferSize:   4096, // 增加读取缓冲区
		WriteBufferSize:  4096, // 增加写入缓冲区
	}

	conn, _, err := dialer.DialContext(ctx, dialURL, reqHeader)
	if err != nil {
		m.retryCount++
		m.logger.Errorf("Connection attempt %d failed: %v", m.retryCount, err)

		if m.shouldRetry() {
			// 不在这里递归调用connect，让调用者处理重试逻辑
			return fmt.Errorf("connection failed: %v", err)
		}
		return fmt.Errorf("connect failed after %d retries: %v", m.retryCount, err)
	}

	// 更新连接状态（需要加锁保护）
	m.mux.Lock()
	// 关闭旧连接（如果存在）
	if m.conn != nil {
		m.conn.Close()
	}
	m.conn = conn
	m.retryCount = 0
	m.mux.Unlock()

	// 设置pong处理器（仅当使用标准ping时）
	if m.config.PingMessage == "" {
		conn.SetPongHandler(func(appData string) error {
			m.logger.Debugf("Pong received")
			return nil
		})
	}

	m.logger.Infof("WebSocket connected to %s", dialURL)

	// 执行连接成功后的回调
	if m.afterConnHandler != nil {
		if err := m.afterConnHandler(m); err != nil {
			m.logger.Errorf("After connection handler error: %v", err)
			// 连接回调失败，关闭连接
			conn.Close()
			return fmt.Errorf("after connection handler failed: %v", err)
		}
		m.logger.Debugf("After connection handler executed successfully")
	}
	// 记录连接指标
	m.metrics.IncrementCounter("websocket.connections.established", map[string]string{
		"url": dialURL,
	})

	return nil
}

// shouldRetry 判断是否重试
func (m *Websocket) shouldRetry() bool {
	return m.config.MaxRetries == 0 || m.retryCount < m.config.MaxRetries
}

// listenLoop 监听消息，支持上下文取消
func (m *Websocket) listenLoop() {
	defer func() {
		// 安全地关闭连接
		m.mux.Lock()
		if m.conn != nil {
			m.conn.Close()
			m.conn = nil
		}
		m.mux.Unlock()

		// 检查是否需要重连
		if m.shouldRetry() {
			m.logger.Warnf("Reconnecting... (attempt %d)", m.retryCount+1)
			// 使用延迟重连，避免立即递归
			m.goroutines.Add(1)
			go func() {
				defer m.goroutines.Done()
				time.Sleep(time.Duration(m.config.RetryDelay) * time.Second)
				if err := m.connect(m.dialURL); err == nil {
					// 重新启动监听循环
					m.goroutines.Add(2)
					go func() {
						defer m.goroutines.Done()
						m.listenLoop()
					}()
					go func() {
						defer m.goroutines.Done()
						m.pingLoop()
					}()
				} else {
					m.logger.Errorf("Reconnect failed: %v", err)
				}
			}()
		} else {
			m.logger.Warnf("WebSocket permanently closed after %d retries", m.retryCount)
		}
	}()

	for {
		select {
		case <-m.ctx.Done():
			return
		default:
			// 安全地获取连接
			m.mux.RLock()
			conn := m.conn
			m.mux.RUnlock()

			if conn == nil {
				m.logger.Warnf("Connection is nil, stopping listen loop")
				return
			}

			_, message, err := conn.ReadMessage()
			if err != nil {
				m.logger.Errorf("Read error: %v", err)
				return
			}

			// 记录消息计数（原子操作，无需锁）
			atomic.AddInt64(&m.messageCount, 1)

			m.metrics.IncrementCounter("websocket.messages.received", map[string]string{
				"url": m.dialURL,
			})

			// 异步处理消息，避免阻塞读取循环
			go func() {
				defer func() {
					if r := recover(); r != nil {
						m.logger.Errorf("Handler panic: %v", r)
						m.metrics.IncrementCounter("websocket.handler.panic", map[string]string{
							"url": m.dialURL,
						})
					}
				}()
				m.messageHandler(message)
			}()
		}
	}
}

// pingLoop 心跳循环（支持标准ping/pong和自定义JSON消息）
func (m *Websocket) pingLoop() {
	ticker := time.NewTicker(time.Duration(m.config.PingInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.mux.RLock()
			conn := m.conn
			m.mux.RUnlock()

			if conn == nil {
				return
			}

			// 根据配置选择心跳方式
			if m.config.PingMessage != "" {
				// 发送JSON消息作为心跳
				if err := conn.WriteMessage(websocket.TextMessage, []byte(m.config.PingMessage)); err != nil {
					m.logger.Errorf("Ping message error: %v", err)
					return // 触发重连
				}
				m.logger.Debugf("Ping message sent: %s", m.config.PingMessage)
			} else {
				// 使用标准ping帧
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					m.logger.Errorf("Ping error: %v", err)
					return // 触发重连
				}
			}
		}
	}
}

// IsConnected 检查连接状态
func (m *Websocket) IsConnected() bool {
	m.mux.RLock()
	defer m.mux.RUnlock()
	return m.isRunning && m.conn != nil
}

// GetRetryCount 获取当前重试次数
func (m *Websocket) GetRetryCount() int {
	m.mux.RLock()
	defer m.mux.RUnlock()
	return m.retryCount
}

// GetMessageCount 获取消息总数
func (m *Websocket) GetMessageCount() int64 {
	return atomic.LoadInt64(&m.messageCount)
}

// GetUptime 获取运行时间
func (m *Websocket) GetUptime() time.Duration {
	return time.Since(m.startTime)
}

// GetDialURL 获取拨号URL
func (m *Websocket) GetDialURL() string {
	return m.dialURL
}

// SetDialURL 设置拨号URL
func (m *Websocket) SetDialURL(dialURL string) {
	m.dialURL = dialURL
}

// GetStats 获取统计信息
func (m *Websocket) GetStats() map[string]interface{} {
	m.mux.RLock()
	defer m.mux.RUnlock()

	return map[string]interface{}{
		"is_connected":  m.isRunning && m.conn != nil,
		"retry_count":   m.retryCount,
		"message_count": atomic.LoadInt64(&m.messageCount),
		"uptime":        time.Since(m.startTime).String(),
		"dial_url":      m.dialURL,
	}
}

// Close 关闭连接
func (m *Websocket) Close() {
	m.mux.Lock()
	if !m.isRunning {
		m.mux.Unlock()
		return
	}

	// 取消上下文，停止所有goroutine
	m.cancel()
	m.isRunning = false
	m.mux.Unlock()

	if m.conn != nil {
		// 安全地发送关闭帧
		func() {
			defer func() {
				if r := recover(); r != nil {
					// 忽略关闭时的错误
					m.logger.Debugf("Close frame send panic (ignored): %v", r)
				}
			}()
			// 设置关闭超时
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// 尝试发送关闭帧
			select {
			case <-ctx.Done():
				m.logger.Warnf("Close frame timeout")
			default:
				m.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			}
		}()

		// 关闭连接
		m.conn.Close()
		m.conn = nil
	}

	// 等待所有goroutine完成
	m.goroutines.Wait()
	m.logger.Infof("WebSocket connection closed")
}
