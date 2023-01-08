package logger

import (
	"context"
	"strings"
)

// Config for logger
type Config struct {
	// Log is a key property that will change the logging mode.
	// Use "Dev" to enable development mode.
	Log string

	// Level is the minimum enabled logging level.
	// Messages with a lower level will be discarded.
	// DebugLevel will be used by default.
	// Use function LevelFromString to set it up using
	// the level string representation.
	Level Level

	// OutputPaths can be used to defined the logger
	// output channels. "stdout" by default.
	OutputPaths []string

	// CtxMiddlewares an arbitrary number of
	// custom context middleware to run when
	// logging an entry with context.
	CtxMiddlewares []CtxMiddleware

	// SkipDefaultMiddlewares when true it skip
	// adding the default ctx middlewares.
	SkipDefaultMiddlewares bool

	// DisableStacktrace when true the stack
	// trace won't be added to log entries
	// above info level
	DisableStacktrace bool
}

// CtxMiddleware is a middleware that will be executed every time
// a context is passed to the logger. It can return an arbitrary number
// of fields that will added to the logger.
type CtxMiddleware func(context.Context) []interface{}

// Level indicates the log entry level.
type Level int

// Available log levels
const (
	DebugLevel Level = iota
	InfoLevel
	WarningLevel
	ErrorLevel
	PanicLevel
	FatalLevel
)

var levelNames = []string{"debug", "info", "warning", "error", "panic", "fatal"}

// String return the string representation of a log level.
func (l Level) String() string {
	return levelNames[l]
}

// LevelFromString returns the logger level according to the
// given string representation, the level match will be evaluated
// as case insensitive.
//
// If the level is unknown it will return DebugLevel.
func LevelFromString(level string) Level {
	switch strings.ToLower(level) {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warning":
		return WarningLevel
	case "error":
		return ErrorLevel
	case "panic":
		return PanicLevel
	case "fatal":
		return FatalLevel
	default:
		return DebugLevel
	}
}

// DefaultMiddlewares the default middlewares that will be used on new loggers.
var DefaultMiddlewares = []CtxMiddleware{RequestIDMiddleware}

// Logger can write log entries using different writer.
type Logger struct {
	writer         Writer
	level          Level
	ctxMiddlewares []CtxMiddleware
}

// New creates a new logger with the default writer.
func New(cfg Config) (Logger, error) {
	w, err := newZapLogger(cfg, 2)
	if err != nil {
		return Logger{}, err
	}

	return NewWithWriter(cfg, w), nil
}

// NewWithWriter creates a new logger with a specific writer.
func NewWithWriter(cfg Config, writer Writer) Logger {
	l := Logger{
		writer:         writer,
		ctxMiddlewares: cfg.CtxMiddlewares,
		level:          cfg.Level,
	}
	if cfg.SkipDefaultMiddlewares {
		return l
	}

	l.ctxMiddlewares = append(l.ctxMiddlewares, DefaultMiddlewares...)
	return l
}

// Must wraps logger constructors and panic if error occurred.
// Usage: lg := logger.Must(logger.New(cfg))
func Must(l Logger, err error) Logger {
	if err != nil {
		panic(err)
	}
	return l
}

// Debug logs a debug message.
func (l Logger) Debug(args ...interface{}) {
	l.Log(DebugLevel, args...)
}

// Debugf logs a debug message indicating a printf compatible format.
func (l Logger) Debugf(str string, args ...interface{}) {
	l.Logf(DebugLevel, str, args...)
}

// Info logs an info message.
func (l Logger) Info(args ...interface{}) {
	l.Log(InfoLevel, args...)
}

// Infof logs an info message indicating a printf compatible format.
func (l Logger) Infof(str string, args ...interface{}) {
	l.Logf(InfoLevel, str, args...)
}

// Warn logs a warning message.
func (l Logger) Warn(args ...interface{}) {
	l.Log(WarningLevel, args...)
}

// Warnf logs a warning message indicating a printf compatible format.
func (l Logger) Warnf(str string, args ...interface{}) {
	l.Logf(WarningLevel, str, args...)
}

// Info logs an error message.
func (l Logger) Error(args ...interface{}) {
	l.Log(ErrorLevel, args...)
}

// Errorf logs an error message indicating a printf compatible format.
func (l Logger) Errorf(str string, args ...interface{}) {
	l.Logf(ErrorLevel, str, args...)
}

// Panic logs an panic Level message and triggers a panic.
func (l Logger) Panic(args ...interface{}) {
	l.Log(PanicLevel, args...)
}

// Panicf logs a panic message indicating a printf compatible format
// and triggers a panic.
func (l Logger) Panicf(str string, args ...interface{}) {
	l.Logf(PanicLevel, str, args...)
}

// Fatal logs an fatal Level message and terminate the execution.
func (l Logger) Fatal(args ...interface{}) {
	l.Log(FatalLevel, args...)
}

// Fatalf logs a fatal message indicating a printf compatible format
// and terminate the execution.
func (l Logger) Fatalf(str string, args ...interface{}) {
	l.Logf(FatalLevel, str, args...)
}

// Log logs a message
func (l Logger) Log(level Level, args ...interface{}) {
	if level < l.level {
		return
	}
	l.innerWriter().Log(level, args...)
}

// Logf logs a message indicating a printf compatible format
func (l Logger) Logf(level Level, str string, args ...interface{}) {
	if level < l.level {
		return
	}
	l.innerWriter().Logf(level, str, args...)
}

// Cond logs a message with a different log level depending on the given condition
// Usage example:
// res, err := operationX()
// logger.Cond(err == nil, logger.DebugLevel, logger.ErrorLevel, "operation X completed", "result", res)
func (l Logger) Cond(condition bool, trueLvl, falseLvl Level, args ...interface{}) {
	l.Log(conditional(condition, trueLvl, falseLvl), args...)
}

// Condf logs a message with a conditional log level indicating a printf compatible format
// See `Cond` for usage
func (l Logger) Condf(condition bool, trueLvl, falseLvl Level, str string, args ...interface{}) {
	l.Logf(conditional(condition, trueLvl, falseLvl), str, args...)
}

// With returns a new logger with fields that will be add to every log entry.
func (l Logger) With(fields ...interface{}) Logger {
	return l.clone(l.innerWriter().With(fields...))
}

// WithMiddleware returns a new logger with more middlewares
func (l Logger) WithMiddleware(middlewares ...CtxMiddleware) Logger {
	cp := l.clone(l.innerWriter())
	cp.ctxMiddlewares = append(cp.ctxMiddlewares, middlewares...)
	return cp
}

// WithContext returns a new logger adding the fields that may be extracted
// from the given context.
func (l Logger) WithContext(ctx context.Context) Logger {
	var fields []interface{}
	for _, m := range l.ctxMiddlewares {
		if fs := m(ctx); len(fs) > 0 {
			fields = append(fields, fs...)
		}
	}
	if len(fields) == 0 {
		return l
	}
	return l.With(fields...)
}

// WithError adds an error as a log field.
func (l Logger) WithError(err error) Logger {
	return l.With("error", err)
}

// Sync ensures that all log entries are written.
func (l Logger) Sync() {
	l.innerWriter().Sync()
}

// noOpWriter is a singleton writer that all zero-value
// loggers can use safely concurrently
var noOpWriter = newNoOpLogger()

func (l Logger) innerWriter() Writer {
	if l.writer == nil {
		return noOpWriter
	}
	return l.writer
}

func (l *Logger) clone(w Writer) Logger {
	return Logger{
		writer:         w,
		ctxMiddlewares: l.ctxMiddlewares,
		level:          l.level,
	}
}

// Writer interface allows writing log entries.
type Writer interface {
	With(fields ...interface{}) Writer
	Log(level Level, args ...interface{})
	Logf(level Level, str string, args ...interface{})
	Sync()
}

func conditional(condition bool, trueLvl, falseLvl Level) Level {
	if !condition {
		return falseLvl
	}
	return trueLvl
}
