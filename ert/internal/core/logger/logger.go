package logger

import (
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

	var writes []zapcore.WriteSyncer
	if logFile != "" {
		dir := filepath.Dir(logFile)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}

		fileWriter := zapcore.NewMutexWriter(zapcore.AddSync(&zapcore.BufferedWriteSyncer{
			WS: zapcore.AddSync(&zapcore.MultiWriteSyncer{
				zapcore.AddSync(os.Stdout),
				zapcore.AddSync(&rotateFile{
					filename:   logFile,
					maxSize:    int64(maxSize) * 1024 * 1024,
					maxBackups: maxBackups,
					maxAge:     maxAge,
					compress:   compress,
				}),
			}),
			FlushInterval: time.Second * 5,
		}))
		writes = append(writes, fileWriter)
	} else {
		writes = append(writes, zapcore.AddSync(os.Stdout))
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.NewMultiWriteSyncer(writes...),
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

func Debug(msg string, fields ...zap.Field) {
	GetLogger().Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	GetLogger().Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	GetLogger().Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	GetLogger().Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	GetLogger().Fatal(msg, fields...)
}

type rotateFile struct {
	filename   string
	maxSize    int64
	maxBackups int
	maxAge     int
	compress   bool
	current    int64
	file       *os.File
}

func (rf *rotateFile) Write(p []byte) (n int, err error) {
	if rf.file == nil {
		rf.file, err = os.OpenFile(rf.filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return 0, err
		}
		info, _ := rf.file.Stat()
		rf.current = info.Size()
	}

	if rf.current >= rf.maxSize {
		rf.file.Close()
		rf.rotate()
	}

	n, err = rf.file.Write(p)
	rf.current += int64(n)
	return
}

func (rf *rotateFile) rotate() error {
	backfile := rf.filename + ".1"
	os.Rename(rf.filename, backfile)
	rf.file, _ = os.Create(rf.filename)
	rf.current = 0
	return nil
}

func (rf *rotateFile) Sync() error {
	if rf.file != nil {
		return rf.file.Sync()
	}
	return nil
}
