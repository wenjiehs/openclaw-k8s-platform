package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger 日志记录器封装
type Logger struct {
	*zap.Logger
}

// NewLogger 创建日志记录器
// mode: "debug" 使用开发模式（彩色、人类可读），"release" 使用生产模式（JSON）
func NewLogger(mode string) (*Logger, error) {
	var zapLogger *zap.Logger
	var err error

	if mode == "release" {
		// 生产模式：JSON 格式，便于日志聚合系统解析
		cfg := zap.NewProductionConfig()
		cfg.EncoderConfig.TimeKey = "timestamp"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		zapLogger, err = cfg.Build()
	} else {
		// 开发模式：彩色输出，便于调试
		cfg := zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		zapLogger, err = cfg.Build()
	}

	if err != nil {
		return nil, err
	}

	return &Logger{zapLogger}, nil
}

// Info 输出 Info 级别日志
func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.Logger.Info(msg, fields...)
}

// Warn 输出 Warn 级别日志
func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.Logger.Warn(msg, fields...)
}

// Error 输出 Error 级别日志
func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.Logger.Error(msg, fields...)
}

// Fatal 输出 Fatal 级别日志并终止程序
func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.Logger.Fatal(msg, fields...)
}

// Debug 输出 Debug 级别日志
func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.Logger.Debug(msg, fields...)
}

// With 创建带有额外字段的子日志记录器
func (l *Logger) With(fields ...zap.Field) *Logger {
	return &Logger{l.Logger.With(fields...)}
}
