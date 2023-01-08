package logger

import (
	"bytes"
	"fmt"
)

// Recorder is a writer that will record all the log
// entries generated.
// It is useful for checking that the expected entries
// are being logged.
type Recorder struct {
	fields     []interface{}
	syncCalled bool

	parent  *Recorder
	entries []LogEntry
}

// LogEntry is holds a single log entry information.
type LogEntry struct {
	Level  Level
	Str    string
	Args   []interface{}
	Fields []interface{}
}

// With return a new recorder with custom fields added.
func (rec *Recorder) With(fields ...interface{}) Writer {
	var all []interface{}
	all = append(all, rec.fields...)
	all = append(all, fields...)
	return rec.clone(all)
}

// Log records a new log entry
func (rec *Recorder) Log(level Level, args ...interface{}) {
	rec.record(level, "", args...)
}

// Logf records a new printf compatible log entry
func (rec *Recorder) Logf(level Level, str string, args ...interface{}) {
	rec.record(level, str, args...)
}

// Sync signal the recorder that the sync operation has been triggered.
func (rec *Recorder) Sync() {
	rec.top().syncCalled = true
}

// SyncCalled returns if the Sync operation was called.
func (rec *Recorder) SyncCalled() bool {
	return rec.top().syncCalled
}

// Entries returns the recorded log entries.
func (rec *Recorder) Entries() []LogEntry {
	return rec.top().entries
}

// Dump will dump all the entries.
func (rec *Recorder) Dump() []byte {
	var b bytes.Buffer
	for _, e := range rec.Entries() {
		b.WriteByte('[')
		b.WriteString(fmt.Sprintf("%-7s", e.Level.String()))
		b.WriteByte(']')

		if e.Str != "" {
			b.WriteByte(' ')
			b.WriteString(fmt.Sprintf(e.Str, e.Args...))
		}
		if e.Str == "" && len(e.Args) > 0 {
			b.WriteByte(' ')
			b.WriteByte('[')
			for i, a := range e.Args {
				if i > 0 {
					b.WriteByte(',')
					b.WriteByte(' ')
				}
				b.WriteString(fmt.Sprint(a))
			}
			b.WriteByte(']')
		}

		b.WriteByte(' ')
		b.WriteByte('{')
		for i, f := range e.Fields {
			if i > 0 {
				b.WriteByte(',')
				b.WriteByte(' ')
			}
			b.WriteString(fmt.Sprint(f))
		}
		b.WriteByte('}')
		b.WriteByte('\n')
	}

	return b.Bytes()
}

// top will get the top-most recorder.
func (rec *Recorder) top() *Recorder {
	var (
		top    = rec
		parent = rec.parent
	)
	for parent != nil {
		top = parent
		parent = top.parent
	}
	return top
}

func (rec *Recorder) record(level Level, str string, args ...interface{}) {
	var top = rec.top()
	e := LogEntry{
		Level:  level,
		Str:    str,
		Args:   args,
		Fields: make([]interface{}, len(rec.fields)),
	}
	copy(e.Fields, rec.fields)
	top.entries = append(top.entries, e)
}

func (rec *Recorder) clone(fields []interface{}) *Recorder {
	cp := Recorder{
		parent: rec,
	}
	cp.fields = append(cp.fields, fields...)
	return &cp
}
