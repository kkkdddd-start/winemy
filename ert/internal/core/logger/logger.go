package logger

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var globalLogger *zap.Logger

func Init(level, logFile string, maxSize, maxBackups, maxAge int, compress bool) error {
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var writeSyncer zapcore.WriteSyncer
	if logFile != "" {
		dir := filepath.Dir(logFile)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return err
		}
		writeSyncer = zapcore.AddSync(file)
	} else {
		writeSyncer = zapcore.AddSync(os.Stdout)
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		writeSyncer,
		zapLevel,
	)

	globalLogger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	return nil
}

func GetLogger() *zap.Logger {
	if globalLogger == nil {
		globalLogger, _ = zap.NewProduction()
	}
	return globalLogger
}

func Debug(msg string) {
	GetLogger().Debug(msg)
}

func Info(msg string) {
	GetLogger().Info(msg)
}

func Warn(msg string) {
	GetLogger().Warn(msg)
}

func Error(msg string) {
	GetLogger().Error(msg)
}

func Fatal(msg string) {
	GetLogger().Fatal(msg)
}

func Errorf(format string, args ...interface{}) {
	GetLogger().Error(fmt.Sprintf(format, args...))
}

func Infof(format string, args ...interface{}) {
	GetLogger().Info(fmt.Sprintf(format, args...))
}

func Debugf(format string, args ...interface{}) {
	GetLogger().Debug(fmt.Sprintf(format, args...))
}

func Warnf(format string, args ...interface{}) {
	GetLogger().Warn(fmt.Sprintf(format, args...))
}

func Fatalf(format string, args ...interface{}) {
	GetLogger().Fatal(fmt.Sprintf(format, args...))
}
