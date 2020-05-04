package wlog

import (
	"fmt"
	"os"
)

// scopedLogger implements interface Logger. This type is used when it is necessary
// to have a separate scope where it's possible to add fields without interfering with the
// parent Logger instance
type scopedLogger struct {
	logger *logger
	fields Fields
}

// GetLogLevel implements Logger.GetLogLevel
func (s *scopedLogger) GetLogLevel() LogLevel {
	return s.logger.logLevel
}

// GetFields implements Logger.GetFields
func (s *scopedLogger) GetFields() Fields {
	return s.fields
}

// GetFieldMapping implements Logger.GetFields
func (s *scopedLogger) GetFieldMapping() FieldMapping {
	return s.logger.fieldMapping
}

// Debugf formats and logs a debug message
func (s *scopedLogger) Debugf(format string, v ...interface{}) {
	if Dbg < s.GetLogLevel() {
		return
	}
	s.logger.writeWithFields(Dbg, fmt.Sprintf(format, v...), s.fields, s.GetFieldMapping())
}

// Debug logs a debug message
func (s *scopedLogger) Debug(v ...interface{}) {
	if Dbg < s.GetLogLevel() {
		return
	}
	s.logger.writeWithFields(Dbg, fmt.Sprint(v...), s.fields, s.GetFieldMapping())
}

// Infof formats and logs an informal message
func (s *scopedLogger) Infof(format string, v ...interface{}) {
	s.logger.writeWithFields(Nfo, fmt.Sprintf(format, v...), s.fields, s.GetFieldMapping())
}

// Info logs an informal message
func (s *scopedLogger) Info(v ...interface{}) {
	s.logger.writeWithFields(Nfo, fmt.Sprint(v...), s.fields, s.GetFieldMapping())
}

// Warningf formats and logs a warning message
func (s *scopedLogger) Warningf(format string, v ...interface{}) {
	s.logger.writeWithFields(Wrn, fmt.Sprintf(format, v...), s.fields, s.GetFieldMapping())
}

// Warning logs a warning message
func (s *scopedLogger) Warning(v ...interface{}) {
	s.logger.writeWithFields(Wrn, fmt.Sprint(v...), s.fields, s.GetFieldMapping())
}

// Errorf formats and logs an error message
func (s *scopedLogger) Errorf(format string, v ...interface{}) {
	s.logger.writeWithFields(Err, fmt.Sprintf(format, v...), s.fields, s.GetFieldMapping())
}

// Error logs an error message
func (s *scopedLogger) Error(v ...interface{}) {
	s.logger.writeWithFields(Err, fmt.Sprint(v...), s.fields, s.GetFieldMapping())
}

// Fatalf formats and logs an unrecoverable error message
func (s *scopedLogger) Fatalf(format string, v ...interface{}) {
	s.logger.writeWithFields(Ftl, fmt.Sprintf(format, v...), s.fields, s.GetFieldMapping())
	os.Exit(1)
}

// Fatal logs an unrecoverable error message
func (s *scopedLogger) Fatal(v ...interface{}) {
	s.logger.writeWithFields(Ftl, fmt.Sprint(v...), s.fields, s.GetFieldMapping())
	os.Exit(1)
}

// GetFormatter gets the writer of the logger
func (s *scopedLogger) GetFormatter() Formatter {
	return s.logger.GetFormatter()
}

// WithScope returns a new instance of Logger based on this Logger.
// Any fields from this Logger will be included to the new
// scoped Logger instance
func (s *scopedLogger) WithScope(fields Fields) Logger {

	// Input should not be touched. Make a value copy
	scopeFields := Fields{}

	// Copy scoped fields from this instance
	for k, v := range s.fields {
		scopeFields[k] = v
	}

	// Copy new scoped fields
	for k, v := range fields {
		scopeFields[k] = v
	}

	return &scopedLogger{s.logger, scopeFields}
}
