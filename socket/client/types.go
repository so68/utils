package client

import (
	"fmt"
)

// Config WebSocket 配置
type Config struct {
	MaxRetries   int               // 最大重试次数，0=无限
	RetryDelay   int               // 重试间隔（秒）
	PingInterval int               // 心跳间隔（秒）
	PingTimeout  int               // 心跳超时（秒）
	PingMessage  string            // 心跳消息（JSON格式），为空则使用标准ping帧
	Headers      map[string]string // 自定义请求头
}

// DefaultConfig 返回默认配置
func DefaultConfig() Config {
	return Config{
		MaxRetries:   5,  // 默认最多重试5次
		RetryDelay:   5,  // 默认重试间隔5秒
		PingInterval: 30, // 默认心跳间隔30秒
		PingTimeout:  10, // 默认心跳超时10秒
	}
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.RetryDelay < 0 {
		return fmt.Errorf("RetryDelay must be non-negative")
	}
	if c.PingInterval <= 0 {
		return fmt.Errorf("PingInterval must be positive")
	}
	if c.PingTimeout <= 0 {
		return fmt.Errorf("PingTimeout must be positive")
	}
	return nil
}

// Metrics 性能指标接口
type Metrics interface {
	IncrementCounter(name string, tags map[string]string)
	RecordHistogram(name string, value float64, tags map[string]string)
	RecordGauge(name string, value float64, tags map[string]string)
}

// NoopMetrics 空指标实现
type NoopMetrics struct{}

func (n *NoopMetrics) IncrementCounter(_ string, _ map[string]string)           {}
func (n *NoopMetrics) RecordHistogram(_ string, _ float64, _ map[string]string) {}
func (n *NoopMetrics) RecordGauge(_ string, _ float64, _ map[string]string)     {}
