package wlog

import (
	"bytes"
	"fmt"
	"os"
	"sync"
	"time"
)

// ScopedLogger is a lighter version of Logger. It implements all the
// logging methods and the interface WLogger. This type is used when it is necessary
// to have a separate scope where it is possible to add fields without interfering with the
// global instance of Logger. ScopedLogger instances are created  by calling wlog.WithScope
type ScopedLogger struct {
	fields    Fields
	logger    *Logger
	formatter Formatter
	mutex     sync.Mutex
}

func (s *ScopedLogger) lock() {
	s.mutex.Lock()
}

func (s *ScopedLogger) unlock() {
	s.mutex.Unlock()
}

// GetFields implements WLogger.GetFields
func (s *ScopedLogger) GetFields() Fields {
	return s.fields
}

// GetLogLevel implements WLogger.GetLogLevel
func (s *ScopedLogger) GetLogLevel() LogLevel {
	return s.logger.logLevel
}

// Debugf formats and logs a debug message
func (s *ScopedLogger) Debugf(format string, v ...interface{}) {
	if Dbg < s.GetLogLevel() {
		return
	}
	s.writef(Dbg, fmt.Sprintf(format, v...))
}

// Debug logs a debug message
func (s *ScopedLogger) Debug(v ...interface{}) {
	if Dbg < s.GetLogLevel() {
		return
	}
	s.writef(Dbg, fmt.Sprint(v...))
}

// Infof formats and logs an informal message
func (s *ScopedLogger) Infof(format string, v ...interface{}) {
	s.writef(Nfo, fmt.Sprintf(format, v...))
}

// Info logs an informal message
func (s *ScopedLogger) Info(v ...interface{}) {
	s.writef(Nfo, fmt.Sprint(v...))
}

// Warningf formats and logs a warning message
func (s *ScopedLogger) Warningf(format string, v ...interface{}) {
	s.writef(Wrn, fmt.Sprintf(format, v...))
}

// Warning logs a warning message
func (s *ScopedLogger) Warning(v ...interface{}) {
	s.writef(Wrn, fmt.Sprint(v...))
}

// Errorf formats and logs an error message
func (s *ScopedLogger) Errorf(format string, v ...interface{}) {
	s.writef(Err, fmt.Sprintf(format, v...))
}

// Error logs an error message
func (s *ScopedLogger) Error(v ...interface{}) {
	s.writef(Err, fmt.Sprint(v...))
}

// Fatalf formats and logs an unrecoverable error message
func (s *ScopedLogger) Fatalf(format string, v ...interface{}) {
	s.writef(Ftl, fmt.Sprintf(format, v...))
	os.Exit(1)
}

// Fatal logs an unrecoverable error message
func (s *ScopedLogger) Fatal(v ...interface{}) {
	s.writef(Ftl, fmt.Sprint(v...))
	os.Exit(1)
}

func (s *ScopedLogger) writef(logLevel LogLevel, msg string) {
	timestamp := time.Now()

	entryBuffer := bufferPool.Get().(*bytes.Buffer)
	defer bufferPool.Put(entryBuffer)

	entryBuffer.Reset()

	if err := s.formatter.Format(entryBuffer, logLevel, s, msg, timestamp); err != nil {
		fmt.Fprintf(os.Stderr, "error formatting the log entry: %v", err)
	}

	write(s.logger, entryBuffer, logLevel, msg, timestamp)
}
