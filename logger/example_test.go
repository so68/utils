package logger

import (
	"testing"
	"time"
)

/*
日志库功能演示测试

本文件用于演示和测试日志库的各种功能特性，
包括不同输出格式、日志级别、属性记录等。

运行命令：
go test -v ./logger -run "^TestExample_Console$"
go test -v ./logger -run "^TestExample_ConsoleWithFile$"

测试内容：
1. 不同日志级别的显示效果
2. 带属性的消息显示效果
3. 带调用位置信息的显示效果
4. 时间格式的显示效果
5. 颜色显示效果
6. 混合输出模式测试
*/

// 测试控制台显示效果
func TestExample_Console(t *testing.T) {
	// 配置用于控制台显示
	config := DefaultConfig()
	config.Level = LevelDebug
	config.Output = OutputStdout

	// 使用 NewLogger 创建 logger 实例
	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Debug 级别
	logger.Debug("这是一条调试信息 - 用于开发阶段的详细调试")
	time.Sleep(200 * time.Millisecond)

	// Info 级别
	logger.Info("这是一条普通信息 - 记录系统正常运行状态")
	time.Sleep(200 * time.Millisecond)

	// Warn 级别
	logger.Warn("这是一条警告信息 - 需要注意但不影响系统运行")
	time.Sleep(200 * time.Millisecond)

	// Error 级别
	logger.Error("这是一条错误信息 - 系统发生了需要处理的错误")
	time.Sleep(200 * time.Millisecond)

	// 测试带属性的消息
	logger.Info("这是一条多属性的测试消息",
		"user", "张三",
		"age", 25,
		"city", "北京",
		"vip", true,
		"login_time", time.Now().Format("2006-01-02 15:04:05"),
		"device", "iPhone 15 Pro",
		"browser", "Safari 17.0",
	)
}

// 测试混合输出：控制台文本格式 + 文件JSON格式
func TestExample_ConsoleWithFile(t *testing.T) {
	// 测试混合输出：控制台文本格式 + 文件JSON格式
	config := DefaultConfig()
	config.Level = LevelInfo
	config.Output = OutputBoth
	config.File.Path = "../logs/app.log"

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Debug 级别
	logger.Debug("这是一条调试信息 - 用于开发阶段的详细调试")
	time.Sleep(200 * time.Millisecond)

	// Info 级别
	logger.Info("这是一条普通信息 - 记录系统正常运行状态")
	time.Sleep(200 * time.Millisecond)

	// Warn 级别
	logger.Warn("这是一条警告信息 - 需要注意但不影响系统运行")
	time.Sleep(200 * time.Millisecond)

	// Error 级别
	logger.Error("这是一条错误信息 - 系统发生了需要处理的错误")
	time.Sleep(200 * time.Millisecond)

	// 测试带属性的消息
	logger.Info("这是一条多属性的测试消息",
		"user", "张三",
		"age", 25,
		"city", "北京",
		"vip", true,
		"login_time", time.Now().Format("2006-01-02 15:04:05"),
		"device", "iPhone 15 Pro",
		"browser", "Safari 17.0",
	)
}
