package logger

import (
	"fmt"
	"log/slog"
	"time"
)

// LogLevel 日志级别类型
type LogLevel string

// LogOutput 日志输出类型
type LogOutput string

// 预定义的日志级别
const (
	// 预定义的日志级别
	LevelDebug LogLevel = "debug" // 调试级别
	LevelInfo  LogLevel = "info"  // 信息级别
	LevelWarn  LogLevel = "warn"  // 警告级别
	LevelError LogLevel = "error" // 错误级别

	// ANSI 颜色代码
	colorReset  = "\033[0m"  // 重置颜色
	colorRed    = "\033[31m" // 红色
	colorGreen  = "\033[32m" // 绿色
	colorYellow = "\033[33m" // 黄色
	colorBlue   = "\033[34m" // 蓝色

	// 预定义的输出类型
	OutputStdout LogOutput = "stdout" // 标准输出
	OutputFile   LogOutput = "file"   // 文件输出
	OutputBoth   LogOutput = "both"   // 同时输出到标准输出和文件
)

// Config log/slog 配置
type Config struct {
	Level      LogLevel   `yaml:"level"`          // 日志级别
	Output     LogOutput  `yaml:"output"`         // 日志输出目标
	File       FileConfig `yaml:"file,omitempty"` // 文件输出配置
	TimeFormat string     `yaml:"time_format"`    // 时间格式
	CallerSkip int        `yaml:"caller_skip"`    // 调用栈跳过层数
}

// FileConfig 日志文件输出配置
type FileConfig struct {
	Path       string `yaml:"path"`         // 日志文件路径
	MaxSizeMb  int    `yaml:"max_size_mb"`  // 单个日志文件最大大小 (MB)
	MaxBackups int    `yaml:"max_backups"`  // 最多保留的旧日志文件数量
	MaxAgeDays int    `yaml:"max_age_days"` // 日志文件最大保留天数
	Compress   bool   `yaml:"compress"`     // 是否压缩旧的日志文件
	LocalTime  bool   `yaml:"local_time"`   // 是否使用本地时间
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Level:      LevelInfo,
		Output:     OutputBoth,
		TimeFormat: time.RFC3339,
		CallerSkip: 4, // 正确的调用栈跳过层数，指向用户代码所在位置
		File: FileConfig{
			Path:       "logs/app.log",
			MaxSizeMb:  100,
			MaxBackups: 10,
			MaxAgeDays: 30,
			Compress:   true,
			LocalTime:  true,
		},
	}
}

// Validate 验证日志配置
func (c *Config) Validate() error {
	// 验证日志级别
	validLevels := map[LogLevel]bool{
		LevelDebug: true,
		LevelInfo:  true,
		LevelWarn:  true,
		LevelError: true,
	}
	if !validLevels[c.Level] {
		return fmt.Errorf("invalid log level: %s", c.Level)
	}

	// 验证输出目标
	validOutputs := map[LogOutput]bool{
		OutputStdout: true,
		OutputFile:   true,
		OutputBoth:   true,
	}
	if !validOutputs[c.Output] {
		return fmt.Errorf("invalid log output: %s", c.Output)
	}

	// 如果输出到文件，验证文件相关配置
	if c.Output == OutputFile || c.Output == OutputBoth {
		if c.File.Path == "" {
			return fmt.Errorf("log file path is required when output is file")
		}
		if c.File.MaxSizeMb <= 0 {
			return fmt.Errorf("log max size must be greater than 0")
		}
		if c.File.MaxBackups <= 0 {
			return fmt.Errorf("log max backups must be greater than 0")
		}
		if c.File.MaxAgeDays <= 0 {
			return fmt.Errorf("log max age must be greater than 0")
		}
	}

	// 验证时间格式
	if c.TimeFormat == "" {
		c.TimeFormat = time.RFC3339
	}

	// 验证调用栈跳过层数
	if c.CallerSkip < 0 {
		return fmt.Errorf("caller skip must be greater than or equal to 0")
	}

	return nil
}

// ShouldOutputToConsole 返回是否应该输出到控制台
func (c *Config) ShouldOutputToConsole() bool {
	return c.Output == OutputStdout || c.Output == OutputBoth
}

// ShouldOutputToFile 返回是否应该输出到文件
func (c *Config) ShouldOutputToFile() bool {
	return c.Output == OutputFile || c.Output == OutputBoth
}

// ShouldAddSource 根据日志级别决定是否添加调用位置信息
// Debug、Warn、Error 级别显示调用位置，Info 级别不显示
func (c *Config) ShouldAddSource(level slog.Level) bool {
	switch level {
	case slog.LevelDebug, slog.LevelWarn, slog.LevelError:
		return true
	case slog.LevelInfo:
		return false
	default:
		return true // 对于其他级别，默认显示调用位置
	}
}

// SlogLevel 返回 slog.Level
func (c *Config) SlogLevel() slog.Level {
	switch c.Level {
	case LevelDebug:
		return slog.LevelDebug
	case LevelWarn:
		return slog.LevelWarn
	case LevelError:
		return slog.LevelError
	}
	return slog.LevelInfo
}

// GetLevelColor 返回日志级别对应的颜色代码
func (c *Config) GetLevelColor() string {
	switch c.Level {
	case LevelDebug:
		return colorBlue
	case LevelInfo:
		return colorGreen
	case LevelWarn:
		return colorYellow
	case LevelError:
		return colorRed
	}
	return colorReset
}
