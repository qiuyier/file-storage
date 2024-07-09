package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"sync"
	"time"
)

type ILogger interface {
	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)
	Fatalf(format string, args ...any)
	Panicf(format string, args ...any)
	DPanicf(format string, args ...any)
}

type Logger struct {
	appName    string
	logger     *zap.SugaredLogger
	level      zapcore.Level
	once       sync.Once
	outputPath string
}

func NewLogger() *Logger {
	return &Logger{appName: "file-storage", level: zapcore.InfoLevel}
}

func (l *Logger) initLogger() {
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = zapcore.TimeEncoderOfLayout(time.DateTime)

	outputPath := []string{"stdout"}
	if l.outputPath != "" {
		outputPath = append(outputPath, l.outputPath)
	}

	errOutputPath := []string{"stderr"}
	if l.outputPath != "" {
		errOutputPath = append(errOutputPath, l.outputPath)
	}

	config := zap.Config{
		Level:             zap.NewAtomicLevelAt(l.level),
		Development:       false,
		DisableCaller:     false,
		DisableStacktrace: false,
		Sampling:          nil,
		Encoding:          "json",
		EncoderConfig:     encoderCfg,
		OutputPaths:       outputPath,
		ErrorOutputPaths:  errOutputPath,
		InitialFields: map[string]interface{}{
			"app": l.appName,
		},
	}

	l.logger = zap.Must(config.Build()).Sugar()
}

func (l *Logger) log(level zapcore.Level, msg string, fields ...any) {
	l.once.Do(func() {
		l.initLogger()
	})
	defer l.logger.Sync()

	switch level {
	case zapcore.DebugLevel:
		l.logger.Debugf(msg, fields...)
	case zapcore.InfoLevel:
		l.logger.Infof(msg, fields...)
	case zapcore.WarnLevel:
		l.logger.Warnf(msg, fields...)
	case zapcore.ErrorLevel:
		l.logger.Errorf(msg, fields...)
	case zapcore.DPanicLevel:
		l.logger.DPanicf(msg, fields...)
	case zapcore.PanicLevel:
		l.logger.Panicf(msg, fields...)
	case zapcore.FatalLevel:
		l.logger.Fatalf(msg, fields...)
	default:
		l.logger.Infof(msg, fields...)
	}
}

func (l *Logger) SetAppName(name string) {
	l.appName = name
}

func (l *Logger) SetLevel(level zapcore.Level) {
	l.level = level
}

func (l *Logger) SetOutputPath(path string) {
	l.outputPath = path
}

func (l *Logger) Debugf(format string, args ...any) {
	l.log(zap.DebugLevel, format, args...)
}

func (l *Logger) Infof(format string, args ...any) {
	l.log(zap.InfoLevel, format, args...)
}

func (l *Logger) Warnf(format string, args ...any) {
	l.log(zap.WarnLevel, format, args...)
}

func (l *Logger) Errorf(format string, args ...any) {
	l.log(zap.ErrorLevel, format, args...)
}

func (l *Logger) Fatalf(format string, args ...any) {
	l.log(zap.FatalLevel, format, args...)
}

func (l *Logger) Panicf(format string, args ...any) {
	l.log(zap.PanicLevel, format, args...)
}

func (l *Logger) DPanicf(format string, args ...any) {
	l.log(zap.DPanicLevel, format, args...)
}
