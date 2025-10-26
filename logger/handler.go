package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mattn/go-colorable"
	"gopkg.in/natefinch/lumberjack.v2"
)

// customHandler 实现 slog.Handler 接口，用于 JSON 格式输出
type customHandler struct {
	handler    slog.Handler
	opts       *slog.HandlerOptions
	callerSkip int
	config     *Config
}

func newCustomHandler(w io.Writer, opts *slog.HandlerOptions, config *Config) *customHandler {
	// 创建一个不包含 AddSource 的选项，避免重复的调用位置信息
	handlerOpts := *opts
	handlerOpts.AddSource = false

	return &customHandler{
		handler:    slog.NewJSONHandler(w, &handlerOpts),
		opts:       opts,
		callerSkip: config.CallerSkip,
		config:     config,
	}
}

func (h *customHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *customHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &customHandler{
		handler:    h.handler.WithAttrs(attrs),
		opts:       h.opts,
		callerSkip: h.callerSkip,
		config:     h.config,
	}
}

func (h *customHandler) WithGroup(name string) slog.Handler {
	return &customHandler{
		handler:    h.handler.WithGroup(name),
		opts:       h.opts,
		callerSkip: h.callerSkip,
		config:     h.config,
	}
}

func (h *customHandler) Handle(ctx context.Context, r slog.Record) error {
	// 检查是否已经有调用位置信息（由 mixedHandler 添加）
	hasSourceInfo := false
	r.Attrs(func(attr slog.Attr) bool {
		if attr.Key == "file" || attr.Key == "line" || attr.Key == "function" {
			hasSourceInfo = true
			return false // 停止遍历
		}
		return true
	})

	// 如果没有调用位置信息，且需要添加，则添加
	if !hasSourceInfo && h.config.ShouldAddSource(r.Level) {
		// customHandler 单独使用时的调用栈深度：Handle(0) -> slog.log(1) -> slog.Debug(2) -> UserCode(3)
		// 所以使用 CallerSkip 来获取用户代码的调用位置
		if pc, file, line, ok := runtime.Caller(h.callerSkip); ok {
			attrs := []slog.Attr{
				slog.String("file", file),
				slog.Int("line", line),
				slog.String("function", runtime.FuncForPC(pc).Name()),
			}
			r.AddAttrs(attrs...)
		}
	}

	// 直接调用底层的 JSON handler
	return h.handler.Handle(ctx, r)
}

// textHandler 实现 slog.Handler 接口，用于彩色文本格式输出
type textHandler struct {
	handler    slog.Handler
	opts       *slog.HandlerOptions
	callerSkip int
	config     *Config
	workingDir string
	writer     io.Writer
}

func newTextHandler(w io.Writer, opts *slog.HandlerOptions, config *Config) *textHandler {
	workingDir, _ := os.Getwd()
	return &textHandler{
		handler:    slog.NewTextHandler(w, opts),
		opts:       opts,
		callerSkip: config.CallerSkip,
		config:     config,
		workingDir: workingDir,
		writer:     w,
	}
}

func (h *textHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *textHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &textHandler{
		handler:    h.handler.WithAttrs(attrs),
		opts:       h.opts,
		callerSkip: h.callerSkip,
		config:     h.config,
		workingDir: h.workingDir,
		writer:     h.writer,
	}
}

func (h *textHandler) WithGroup(name string) slog.Handler {
	return &textHandler{
		handler:    h.handler.WithGroup(name),
		opts:       h.opts,
		callerSkip: h.callerSkip,
		config:     h.config,
		workingDir: h.workingDir,
		writer:     h.writer,
	}
}

func (h *textHandler) Handle(ctx context.Context, r slog.Record) error {
	// 检查是否已经有调用位置信息（由 mixedHandler 添加）
	hasSourceInfo := false
	r.Attrs(func(attr slog.Attr) bool {
		if attr.Key == "file" || attr.Key == "line" || attr.Key == "function" {
			hasSourceInfo = true
			return false // 停止遍历
		}
		return true
	})

	// 如果没有调用位置信息，且需要添加，则添加
	if !hasSourceInfo && h.config.ShouldAddSource(r.Level) {
		// textHandler 单独使用时的调用栈深度：Handle(0) -> slog.log(1) -> slog.Debug(2) -> UserCode(3)
		// 所以使用 CallerSkip 来获取用户代码的调用位置
		if pc, file, line, ok := runtime.Caller(h.callerSkip); ok {
			attrs := []slog.Attr{
				slog.String("file", file),
				slog.Int("line", line),
				slog.String("function", runtime.FuncForPC(pc).Name()),
			}
			r.AddAttrs(attrs...)
		}
	}

	var builder strings.Builder

	// 时间戳
	timestamp := r.Time.Format(h.config.TimeFormat)
	builder.WriteString(timestamp)
	builder.WriteString(" ")

	// 日志级别（带颜色）
	levelColor := h.getLevelColor(r.Level)
	builder.WriteString(levelColor)
	builder.WriteString("[")
	builder.WriteString(r.Level.String())
	builder.WriteString("]")
	builder.WriteString(colorReset)
	builder.WriteString(" ")

	// 消息
	builder.WriteString(r.Message)

	// 属性（包括调用位置信息）
	h.writeAttrs(&builder, r)

	// 输出
	_, err := fmt.Fprintln(h.writer, builder.String())
	return err
}

// writeAttrs 写入属性信息
func (h *textHandler) writeAttrs(builder *strings.Builder, r slog.Record) {
	first := true
	var sourceAttrs []slog.Attr

	r.Attrs(func(attr slog.Attr) bool {
		// 检查是否是调用位置相关的属性
		if attr.Key == "file" || attr.Key == "line" || attr.Key == "function" {
			sourceAttrs = append(sourceAttrs, attr)
			return true
		}

		// 处理普通属性
		if first {
			builder.WriteString(" | ")
			first = false
		} else {
			builder.WriteString(" | ")
		}
		builder.WriteString(attr.Key)
		builder.WriteString("=")
		builder.WriteString(fmt.Sprintf("%v", attr.Value.Any()))
		return true
	})

	// 如果有调用位置信息，特殊处理
	if len(sourceAttrs) > 0 {
		h.writeSourceFromAttrs(builder, sourceAttrs)
	}
}

// writeSourceFromAttrs 从属性中写入调用位置信息
func (h *textHandler) writeSourceFromAttrs(builder *strings.Builder, sourceAttrs []slog.Attr) {
	var file, line, function string

	// 提取调用位置信息
	for _, attr := range sourceAttrs {
		switch attr.Key {
		case "file":
			file = attr.Value.String()
		case "line":
			line = fmt.Sprintf("%d", attr.Value.Int64())
		case "function":
			function = attr.Value.String()
		}
	}

	if file != "" && line != "" {
		// 相对路径
		relFile := file
		if h.workingDir != "" {
			if rel, err := filepath.Rel(h.workingDir, file); err == nil {
				relFile = rel
			}
		}

		builder.WriteString(" | ")
		builder.WriteString(relFile)
		builder.WriteString(":")
		builder.WriteString(line)
		if function != "" {
			builder.WriteString(" | ")
			builder.WriteString(function)
		}
	}
}

// getLevelColor 根据日志级别返回对应的颜色代码
func (h *textHandler) getLevelColor(level slog.Level) string {
	switch level {
	case slog.LevelDebug:
		return colorBlue
	case slog.LevelInfo:
		return colorGreen
	case slog.LevelWarn:
		return colorYellow
	case slog.LevelError:
		return colorRed
	default:
		return colorReset
	}
}

// mixedHandler 混合处理器，同时支持控制台（文本格式）和文件（JSON格式）输出
type mixedHandler struct {
	consoleHandler slog.Handler // 控制台处理器（文本格式）
	fileHandler    slog.Handler // 文件处理器（JSON格式）
	opts           *slog.HandlerOptions
	config         *Config
}

func newMixedHandler(config *Config, opts *slog.HandlerOptions) *mixedHandler {
	// 创建控制台处理器（文本格式）
	consoleWriter := colorable.NewColorableStdout()
	consoleHandler := newTextHandler(consoleWriter, opts, config)

	// 创建文件处理器（JSON格式）
	fileWriter := &lumberjack.Logger{
		Filename:   config.File.Path,
		MaxSize:    config.File.MaxSizeMb,
		MaxBackups: config.File.MaxBackups,
		MaxAge:     config.File.MaxAgeDays,
		Compress:   config.File.Compress,
		LocalTime:  config.File.LocalTime,
	}
	fileHandler := newCustomHandler(fileWriter, opts, config)

	return &mixedHandler{
		consoleHandler: consoleHandler,
		fileHandler:    fileHandler,
		opts:           opts,
		config:         config,
	}
}

func (h *mixedHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.consoleHandler.Enabled(ctx, level)
}

func (h *mixedHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &mixedHandler{
		consoleHandler: h.consoleHandler.WithAttrs(attrs),
		fileHandler:    h.fileHandler.WithAttrs(attrs),
		opts:           h.opts,
		config:         h.config,
	}
}

func (h *mixedHandler) WithGroup(name string) slog.Handler {
	return &mixedHandler{
		consoleHandler: h.consoleHandler.WithGroup(name),
		fileHandler:    h.fileHandler.WithGroup(name),
		opts:           h.opts,
		config:         h.config,
	}
}

func (h *mixedHandler) Handle(ctx context.Context, r slog.Record) error {
	// 添加调用位置信息到记录中 - 根据日志级别动态决定
	if h.config.ShouldAddSource(r.Level) {
		// mixedHandler 的调用栈深度：Handle(0) -> slog.log(1) -> slog.Debug(2) -> UserCode(3)
		// 所以直接使用 CallerSkip 来获取用户代码的调用位置
		if pc, file, line, ok := runtime.Caller(h.config.CallerSkip); ok {
			attrs := []slog.Attr{
				slog.String("file", file),
				slog.Int("line", line),
				slog.String("function", runtime.FuncForPC(pc).Name()),
			}
			r.AddAttrs(attrs...)
		}
	}

	// 同时处理控制台和文件输出
	var consoleErr, fileErr error

	// 处理控制台输出（文本格式）
	if h.config.ShouldOutputToConsole() {
		consoleErr = h.consoleHandler.Handle(ctx, r)
	}

	// 处理文件输出（JSON格式）
	if h.config.ShouldOutputToFile() {
		fileErr = h.fileHandler.Handle(ctx, r)
	}

	// 如果有错误，返回第一个错误
	if consoleErr != nil {
		return consoleErr
	}
	if fileErr != nil {
		return fileErr
	}

	return nil
}
