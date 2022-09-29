package lock

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LogLevel uint8

const (
	Error = iota + 1
	Warn
	Info
	Debug
)

var level zapcore.Level

type Logger interface {
	Debug(msg string, data ...interface{})
	Info(msg string, data ...interface{})
	Warn(msg string, data ...interface{})
	Error(msg string, data ...interface{})
}

func Zap() (logger *zap.Logger) {
	level = zapcore.DebugLevel
	logger = zap.New(getEncoderCore(), zap.AddStacktrace(level))
	logger = logger.WithOptions(zap.AddCaller(), zap.AddCallerSkip(1))
	return logger
}

// getEncoderConfig 获取zapcore.EncoderConfig
func getEncoderConfig() (config zapcore.EncoderConfig) {
	config = zapcore.EncoderConfig{
		MessageKey: "message",
		LevelKey:   "level",
		TimeKey:    "time",
		NameKey:    "logger",
		CallerKey:  "caller",
		//StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     CustomTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.FullCallerEncoder,
	}
	config.EncodeLevel = zapcore.LowercaseColorLevelEncoder
	return config
}

// getEncoder 获取zapcore.Encoder
func getEncoder() zapcore.Encoder {
	return zapcore.NewConsoleEncoder(getEncoderConfig())
}

// getEncoderCore 获取Encoder的zapcore.Core
func getEncoderCore() (core zapcore.Core) {
	writer := zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout))
	return zapcore.NewCore(getEncoder(), writer, level)
}

// 自定义日志输出时间格式
func CustomTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("[redis-lock] " + "2006/01/02 - 15:04:05.000"))
}

type DefaultLogger struct {
	Level LogLevel
	Log   *zap.Logger
}

func (l *DefaultLogger) Debug(msg string, data ...interface{}) {
	if l.Level >= Debug {
		l.Log.Debug(fmt.Sprintf(msg, data...))
	}
}

func (l *DefaultLogger) Info(msg string, data ...interface{}) {
	if l.Level >= Info {
		l.Log.Info(fmt.Sprintf(msg, data...))
	}
}

func (l *DefaultLogger) Warn(msg string, data ...interface{}) {
	if l.Level >= Warn {
		l.Log.Warn(fmt.Sprintf(msg, data...))
	}
}

func (l *DefaultLogger) Error(msg string, data ...interface{}) {
	if l.Level >= Error {
		l.Log.Error(fmt.Sprintf(msg, data...))
	}
}
