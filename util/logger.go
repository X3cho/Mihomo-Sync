package util

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// LogLevel 日志等级
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARNING
	ERROR
	CRITICAL
)

// ParseLogLevel 解析日志等级字符串
func ParseLogLevel(level string) LogLevel {
	switch level {
	case "DEBUG", "debug":
		return DEBUG
	case "INFO", "info":
		return INFO
	case "WARNING", "warning", "WARN", "warn":
		return WARNING
	case "ERROR", "error":
		return ERROR
	case "CRITICAL", "critical":
		return CRITICAL
	default:
		return INFO
	}
}

// String 日志等级转字符串
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARNING:
		return "WARNING"
	case ERROR:
		return "ERROR"
	case CRITICAL:
		return "CRITICAL"
	default:
		return "INFO"
	}
}

// Logger 日志记录器
type Logger struct {
	mu       sync.Mutex
	file     *os.File
	out      io.Writer
	level    LogLevel
	fileOnly bool
}

// LoggerConfig 日志配置
type LoggerConfig struct {
	File     string
	Level    string
	FileOnly bool
}

// NewLogger 创建日志记录器
func NewLogger(cfg LoggerConfig) (*Logger, error) {
	level := ParseLogLevel(cfg.Level)

	logger := &Logger{
		level:    level,
		fileOnly: cfg.FileOnly,
	}

	// 打开日志文件
	if cfg.File != "" {
		file, err := os.OpenFile(cfg.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("打开日志文件失败：%w", err)
		}
		logger.file = file

		if !cfg.FileOnly {
			logger.out = io.MultiWriter(os.Stdout, file)
		} else {
			logger.out = file
		}
	} else {
		logger.out = os.Stdout
	}

	return logger, nil
}

// log 内部日志方法
func (l *Logger) log(level LogLevel, format string, args ...any) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	_, _ = fmt.Fprintf(l.out, "[%s] %s: ", timestamp, level.String())
	_, _ = fmt.Fprintf(l.out, format+"\n", args...)
}

// Debug 调试日志
func (l *Logger) Debug(format string, args ...any) {
	l.log(DEBUG, format, args...)
}

// Info 信息日志
func (l *Logger) Info(format string, args ...any) {
	l.log(INFO, format, args...)
}

// Warning 警告日志
func (l *Logger) Warning(format string, args ...any) {
	l.log(WARNING, format, args...)
}

// Error 错误日志
func (l *Logger) Error(format string, args ...any) {
	l.log(ERROR, format, args...)
}

// Critical 严重错误日志
func (l *Logger) Critical(format string, args ...any) {
	l.log(CRITICAL, format, args...)
}

// Close 关闭日志文件
func (l *Logger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}
