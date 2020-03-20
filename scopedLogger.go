package wlog

import (
	"fmt"
	"os"
)

// scopedLogger implements interface WLogger. This type is used when it is necessary
// to have a separate scope where it's possible to add fields without interfering with the
// parent Logger instance. scopedLogger instances are created  by calling wlog.WithScope
type scopedLogger struct {
	logger *Logger
	fields Fields
}

// GetLogLevel implements WLogger.GetLogLevel
func (s *scopedLogger) GetLogLevel() LogLevel {
	return s.logger.logLevel
}

// GetFields implements WLogger.GetFields
func (s *scopedLogger) GetFields() Fields {
	return s.fields
}

// Debugf formats and logs a debug message
func (s *scopedLogger) Debugf(format string, v ...interface{}) {
	if Dbg < s.GetLogLevel() {
		return
	}
	s.logger.writeWithFields(Dbg, fmt.Sprintf(format, v...), s.fields)
}

// Debug logs a debug message
func (s *scopedLogger) Debug(v ...interface{}) {
	if Dbg < s.GetLogLevel() {
		return
	}
	s.logger.writeWithFields(Dbg, fmt.Sprint(v...), s.fields)
}

// Infof formats and logs an informal message
func (s *scopedLogger) Infof(format string, v ...interface{}) {
	s.logger.writeWithFields(Nfo, fmt.Sprintf(format, v...), s.fields)
}

// Info logs an informal message
func (s *scopedLogger) Info(v ...interface{}) {
	s.logger.writeWithFields(Nfo, fmt.Sprint(v...), s.fields)
}

// Warningf formats and logs a warning message
func (s *scopedLogger) Warningf(format string, v ...interface{}) {
	s.logger.writeWithFields(Wrn, fmt.Sprintf(format, v...), s.fields)
}

// Warning logs a warning message
func (s *scopedLogger) Warning(v ...interface{}) {
	s.logger.writeWithFields(Wrn, fmt.Sprint(v...), s.fields)
}

// Errorf formats and logs an error message
func (s *scopedLogger) Errorf(format string, v ...interface{}) {
	s.logger.writeWithFields(Err, fmt.Sprintf(format, v...), s.fields)
}

// Error logs an error message
func (s *scopedLogger) Error(v ...interface{}) {
	s.logger.writeWithFields(Err, fmt.Sprint(v...), s.fields)
}

// Fatalf formats and logs an unrecoverable error message
func (s *scopedLogger) Fatalf(format string, v ...interface{}) {
	s.logger.writeWithFields(Ftl, fmt.Sprintf(format, v...), s.fields)
	os.Exit(1)
}

// Fatal logs an unrecoverable error message
func (s *scopedLogger) Fatal(v ...interface{}) {
	s.logger.writeWithFields(Ftl, fmt.Sprint(v...), s.fields)
	os.Exit(1)
}
