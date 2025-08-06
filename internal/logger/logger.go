package logger

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger struct {
	*zap.Logger
	sugar *zap.SugaredLogger
}

type Config struct {
	LogLevel    string
	LogDir      string
	MaxSize     int  // megabytes
	MaxBackups  int  // number of backups
	MaxAge      int  // days
	Compress    bool
	Development bool
}

func New(config Config) (*Logger, error) {
	// Create log directory if it doesn't exist
	if err := os.MkdirAll(config.LogDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Configure log level
	var level zapcore.Level
	switch config.LogLevel {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	// Configure encoder
	var encoderConfig zapcore.EncoderConfig
	if config.Development {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		encoderConfig = zap.NewProductionEncoderConfig()
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	encoder := zapcore.NewJSONEncoder(encoderConfig)

	// Configure file rotation
	fileWriter := &lumberjack.Logger{
		Filename:   filepath.Join(config.LogDir, "rodmcp.log"),
		MaxSize:    config.MaxSize,
		MaxBackups: config.MaxBackups,
		MaxAge:     config.MaxAge,
		Compress:   config.Compress,
	}

	// Create separate log files for different components (for future use)
	_ = &lumberjack.Logger{
		Filename:   filepath.Join(config.LogDir, "mcp.log"),
		MaxSize:    config.MaxSize,
		MaxBackups: config.MaxBackups,
		MaxAge:     config.MaxAge,
		Compress:   config.Compress,
	}

	_ = &lumberjack.Logger{
		Filename:   filepath.Join(config.LogDir, "browser.log"),
		MaxSize:    config.MaxSize,
		MaxBackups: config.MaxBackups,
		MaxAge:     config.MaxAge,
		Compress:   config.Compress,
	}

	// Create multi-writer core
	consoleWriter := zapcore.AddSync(os.Stdout)
	fileCore := zapcore.NewCore(encoder, zapcore.AddSync(fileWriter), level)
	consoleCore := zapcore.NewCore(encoder, consoleWriter, level)
	
	core := zapcore.NewTee(fileCore, consoleCore)

	// Create logger with caller info and stack traces
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return &Logger{
		Logger: logger,
		sugar:  logger.Sugar(),
	}, nil
}

func (l *Logger) Sugar() *zap.SugaredLogger {
	return l.sugar
}

func (l *Logger) WithComponent(component string) *zap.Logger {
	return l.Logger.With(zap.String("component", component))
}

func (l *Logger) WithRequest(requestID string) *zap.Logger {
	return l.Logger.With(zap.String("request_id", requestID))
}

func (l *Logger) LogMCPRequest(method string, params interface{}) {
	l.WithComponent("mcp").Info("MCP request",
		zap.String("method", method),
		zap.Any("params", params),
	)
}

func (l *Logger) LogMCPResponse(method string, result interface{}, err error) {
	if err != nil {
		l.WithComponent("mcp").Error("MCP response error",
			zap.String("method", method),
			zap.Error(err),
		)
	} else {
		l.WithComponent("mcp").Info("MCP response",
			zap.String("method", method),
			zap.Any("result", result),
		)
	}
}

func (l *Logger) LogBrowserAction(action string, url string, duration int64) {
	l.WithComponent("browser").Info("Browser action",
		zap.String("action", action),
		zap.String("url", url),
		zap.Int64("duration_ms", duration),
	)
}

func (l *Logger) LogToolExecution(toolName string, args map[string]interface{}, success bool, duration int64) {
	if success {
		l.WithComponent("tools").Info("Tool execution successful",
			zap.String("tool", toolName),
			zap.Any("args", args),
			zap.Int64("duration_ms", duration),
		)
	} else {
		l.WithComponent("tools").Error("Tool execution failed",
			zap.String("tool", toolName),
			zap.Any("args", args),
			zap.Int64("duration_ms", duration),
		)
	}
}