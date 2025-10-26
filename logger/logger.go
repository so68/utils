package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/mattn/go-colorable"
	"gopkg.in/natefinch/lumberjack.v2"
)

// NewLogger 创建一个新的 logger 实例
func NewLogger(config *Config) (*slog.Logger, error) {
	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// 创建日志目录
	if config.ShouldOutputToFile() {
		if err := os.MkdirAll(filepath.Dir(config.File.Path), 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}
	}

	// 配置输出目标
	var writers []io.Writer

	// 添加文件输出
	if config.ShouldOutputToFile() {
		fileWriter := &lumberjack.Logger{
			Filename:   config.File.Path,
			MaxSize:    config.File.MaxSizeMb,  // MB
			MaxBackups: config.File.MaxBackups, // 保留的旧文件最大数量
			MaxAge:     config.File.MaxAgeDays, // 保留天数
			Compress:   config.File.Compress,   // 是否压缩
			LocalTime:  config.File.LocalTime,  // 是否使用本地时间
		}

		writers = append(writers, fileWriter)
	}

	// 添加控制台输出
	if config.ShouldOutputToConsole() {
		// 使用 colorable 支持 Windows 下的彩色输出
		consoleWriter := colorable.NewColorableStdout()

		writers = append(writers, consoleWriter)
	}

	// 创建自定义的 Handler
	var handler slog.Handler
	handlerOpts := &slog.HandlerOptions{
		Level:     config.SlogLevel(),
		AddSource: false, // 我们使用自定义的动态 AddSource 逻辑
	}

	// 根据输出目标自动选择格式
	if config.ShouldOutputToConsole() && config.ShouldOutputToFile() {
		// 同时输出到控制台和文件，使用混合处理器
		handler = newMixedHandler(config, handlerOpts)
	} else if config.ShouldOutputToConsole() {
		// 只输出到控制台，使用文本格式
		handler = newTextHandler(io.MultiWriter(writers...), handlerOpts, config)
	} else if config.ShouldOutputToFile() {
		// 只输出到文件，使用JSON格式
		handler = newCustomHandler(io.MultiWriter(writers...), handlerOpts, config)
	} else {
		// 默认使用文本格式
		handler = newTextHandler(io.MultiWriter(writers...), handlerOpts, config)
	}

	return slog.New(handler), nil
}
