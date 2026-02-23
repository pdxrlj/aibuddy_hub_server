// Package logger 提供日志记录功能
package logger

import (
	"io"
	"log/slog"
	"os"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger 日志记录器
type Logger struct {
	Output     io.Writer
	FileLogger *FileLogger
	Level      string
}

// FileLogger 文件日志配置
type FileLogger struct {
	Filename   string
	MaxSize    int
	MaxAge     int
	MaxBackups int
	Compress   bool
}

// DefaultFileLogger 返回默认文件日志配置
func DefaultFileLogger() *FileLogger {
	return &FileLogger{
		MaxSize:    100,
		MaxAge:     30,
		MaxBackups: 10,
		Compress:   true,
		Filename:   "logs/app.log",
	}
}

// NewLogger 创建日志记录器
func NewLogger(level string, fileLogger *FileLogger) *Logger {
	return &Logger{
		Level:      level,
		FileLogger: fileLogger,
	}
}

// NewWithOutput 创建带输出目标的日志记录器
func NewWithOutput(output io.Writer) *Logger {
	return &Logger{
		Output: output,
	}
}

// Setup 初始化日志记录器
func (l *Logger) Setup() {
	dsetWriter := io.MultiWriter(os.Stdout)

	if l.FileLogger != nil {
		rotater := lumberjack.Logger{
			Filename:   l.FileLogger.Filename,
			MaxSize:    l.FileLogger.MaxSize,
			MaxAge:     l.FileLogger.MaxAge,
			MaxBackups: l.FileLogger.MaxBackups,
			Compress:   l.FileLogger.Compress,
		}
		dsetWriter = io.MultiWriter(dsetWriter, &rotater)
	}

	if l.Output != nil {
		dsetWriter = io.MultiWriter(dsetWriter, l.Output)
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(dsetWriter,
		&slog.HandlerOptions{
			Level: l.GetLevel(),
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if t, ok := a.Value.Any().(time.Time); ok {
					a.Value = slog.StringValue(t.Format(time.DateTime))
				}

				return a
			},
		})))
}

// GetLevel 获取日志级别
func (l *Logger) GetLevel() slog.Level {
	level := slog.Level(0)
	err := level.UnmarshalText([]byte(l.Level))
	if err != nil {
		slog.Error("Failed to unmarshal log level set default to info", "error", err)
		level = slog.LevelInfo
	}
	return level
}
