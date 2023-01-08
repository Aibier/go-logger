package logger

type noOpLogger struct{}

// NewNoOpLogger creates a logger with the no-op writer
// @deprecated the logger zero-value struct is equivalent.
func NewNoOpLogger() Logger {
	return NewWithWriter(Config{}, newNoOpLogger())
}

// newNoOpLogger ...
func newNoOpLogger() noOpLogger {
	return noOpLogger{}
}

func (z noOpLogger) Log(_ Level, _ ...interface{}) {}

func (z noOpLogger) Logf(_ Level, _ string, _ ...interface{}) {}

func (z noOpLogger) With(_ ...interface{}) Writer {
	return z
}

func (z noOpLogger) Sync() {}
