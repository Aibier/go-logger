package logger

import (
	"os"
	"runtime"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type zapLogger struct {
	logger *zap.SugaredLogger
}

func (z zapLogger) Sync() {
	_ = z.logger.Sync()
}

func (z zapLogger) Log(level Level, args ...interface{}) {
	switch level {
	case DebugLevel:
		z.logger.Debug(args...)
	case InfoLevel:
		z.logger.Info(args...)
	case WarningLevel:
		z.logger.Warn(args...)
	case ErrorLevel:
		z.logger.Error(args...)
	case PanicLevel:
		z.logger.Panic(args...)
	case FatalLevel:
		z.logger.Fatal(args...)
	}
}

func (z zapLogger) Logf(level Level, str string, args ...interface{}) {
	switch level {
	case DebugLevel:
		z.logger.Debugf(str, args...)
	case InfoLevel:
		z.logger.Infof(str, args...)
	case WarningLevel:
		z.logger.Warnf(str, args...)
	case ErrorLevel:
		z.logger.Errorf(str, args...)
	case PanicLevel:
		z.logger.Panicf(str, args...)
	case FatalLevel:
		z.logger.Fatalf(str, args...)
	}
}

func (z zapLogger) With(fields ...interface{}) Writer {
	return zapLogger{logger: z.logger.With(fields...)}
}

// NewZapLogger creates a new logger based on Zap.
// @deprecated use logger.New. keeping this to prevent breaking changes.
func NewZapLogger(conf Config) (Logger, error) {
	return New(conf)
}

// newZapLogger returns a new zap writer.
func newZapLogger(conf Config, callerSkip int) (Writer, error) {
	callerSkip++
	if conf.Log == "Dev" {
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.DisableStacktrace = conf.DisableStacktrace
		if conf.OutputPaths != nil {
			config.OutputPaths = conf.OutputPaths
		}

		logger, err := config.Build()
		if err != nil {
			return nil, err
		}

		return zapLogger{
			logger: logger.WithOptions(zap.AddCallerSkip(callerSkip)).Sugar(),
		}, nil
	}

	initFields := map[string]interface{}{
		"goVersion": runtime.Version(),
		"pid":       os.Getpid(),
	}
	hostname, err := os.Hostname()
	if err == nil {
		initFields["hostname"] = hostname
	}

	outputPaths := conf.OutputPaths
	if outputPaths == nil {
		outputPaths = []string{"stdout"}
	}

	cfg := zap.Config{
		Encoding:          "json",
		Level:             zap.NewAtomicLevelAt(zapcore.DebugLevel),
		OutputPaths:       outputPaths,
		InitialFields:     initFields,
		DisableStacktrace: conf.DisableStacktrace,
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.MillisDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
	}

	logger, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	return zapLogger{
		logger: logger.WithOptions(zap.AddCallerSkip(callerSkip)).Sugar(),
	}, nil
}
