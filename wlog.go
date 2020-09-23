package wlog

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

// Config allows configuration of a logger
type Config struct {
	LogLevel        LogLevel
	Path            string
	TruncateOnStart bool
	StdOut          bool
	Formatter       Formatter
	Writer          io.Writer
}

// LogLevel controls how verbose the output will be
type LogLevel int

// The Log levels available
const (
	Dbg LogLevel = iota
	Nfo
	Wrn
	Err
	Ftl
)

func (l LogLevel) String() string {
	switch l {
	case Dbg:
		return "Debug"
	case Nfo:
		return "Info"
	case Wrn:
		return "Warning"
	case Err:
		return "Error"
	case Ftl:
		return "Fatal"
	}

	return "Unknown"
}

// HookFunc is a callback function triggered
// when a log event occurs
type HookFunc func(time.Time, LogLevel, string)

var (
	// The default logger instance
	defaultLogger *logger
)

func init() {
	defaultLogger = newLogger(nil, Nfo, true)
}

func newLogger(writer io.Writer, logLevel LogLevel, stdOut bool) *logger {
	return &logger{
		writer:       writer,
		logLevel:     logLevel,
		stdOut:       stdOut,
		formatter:    TextFormatter{},
		fields:       Fields{},
		fieldMapping: FieldMapping{"level": "@l", "timestamp": "@t", "message": "@m"},
	}
}

// New creates a new instance of a logger
func New(writer io.Writer, logLevel LogLevel, stdOut bool) MutableLogger {
	return newLogger(
		writer,
		logLevel,
		stdOut,
	)
}

// Logger is the interface that wlog loggers implements
type Logger interface {
	Debugf(format string, v ...interface{})
	Debug(v ...interface{})
	Infof(format string, v ...interface{})
	Info(v ...interface{})
	Warningf(format string, v ...interface{})
	Warning(v ...interface{})
	Errorf(format string, v ...interface{})
	Error(v ...interface{})
	Fatalf(format string, v ...interface{})
	Fatal(v ...interface{})
	GetFields() Fields
	GetFieldMapping() FieldMapping
	GetLogLevel() LogLevel
	GetFormatter() Formatter
	WithScope(fields Fields) Logger
}

// MutableLogger extends the Logger interface by providing
// mutable characteristics to a logger
type MutableLogger interface {
	Logger
	SetFormatter(formatter Formatter)
	SetStdOut(enable bool)
	SetFields(fields Fields)
	SetLogLevel(logLevel LogLevel)
	Configure(cfg *Config)
	SetFieldMapping(fieldMapping FieldMapping)
}

// FieldMapping is used to map field names when using
// JSONFormatter in compact mode
type FieldMapping map[string]string

// Fields is a map containing the fields that will be added to every log entry
type Fields map[string]interface{}

// logger implements the MutableLogger interface
// providing the default logger instance as well
// as the base for new loggers created with New()
type logger struct {
	writer       io.Writer
	logLevel     LogLevel
	stdOut       bool
	mutex        sync.Mutex
	hooks        map[LogLevel][]HookFunc
	fields       Fields
	formatter    Formatter
	fieldMapping FieldMapping
}

var bufferPool = sync.Pool{New: func() interface{} {
	return new(bytes.Buffer)
}}

func (l *logger) lock() {
	l.mutex.Lock()
}

func (l *logger) unlock() {
	l.mutex.Unlock()
}

// SetFieldMapping add custom field mapping for structured log
func (l *logger) SetFieldMapping(fieldMapping FieldMapping) {
	l.lock()
	defer l.unlock()

	var mapping = make(FieldMapping, 3)

	// first add the custom mappings
	for k, v := range fieldMapping {
		if strings.HasPrefix(v, "@") {
			fmt.Fprintf(os.Stderr, "value cannot be prefixed with @: %s", v)
		}
		mapping[k] = v
	}

	// add the default mappings
	for k, v := range l.fieldMapping {
		mapping[k] = v
	}

	l.fieldMapping = mapping

}

// Configure configures a mutable logger
func (l *logger) Configure(cfg *Config) {
	l.SetLogLevel(cfg.LogLevel)
	l.SetStdOut(cfg.StdOut)

	if cfg.Formatter != nil {
		l.SetFormatter(cfg.Formatter)
	}

	if cfg.Path != "" {

		flags := os.O_RDWR | os.O_CREATE

		if cfg.TruncateOnStart {
			flags |= os.O_TRUNC
		} else {
			flags |= os.O_APPEND
		}

		file, err := os.OpenFile(cfg.Path, flags, 0666)
		if err != nil {
			l.Fatal(err)
		}

		l.SetWriter(file)
	} else {
		l.SetWriter(cfg.Writer)
	}
}

// WithScope returns a new instance of Logger. It's fields property
// will contain the fields added to the parent instance of Logger
func (l *logger) WithScope(fields Fields) Logger {

	// Input should not be touched. Make a value copy
	scopeFields := Fields{}

	l.lock()
	defer l.unlock()

	// Copy scoped fields from this instance
	for k, v := range l.fields {
		scopeFields[k] = v
	}

	// Copy new scoped fields
	for k, v := range fields {
		scopeFields[k] = v
	}

	return &scopedLogger{l, scopeFields}
}

// SetGlobalFields set fields in a log instance. These fields will be appended to any
// child scope created with log.WithScope method.
// Deprecated: Please use SetFields instead.
func (l *logger) SetGlobalFields(f Fields) {
	l.SetFields(f)
}

// SetFields set fields for a MutableLogger instance. These fields will be appended to any
// child scope created with log.WithScope method.
func (l *logger) SetFields(f Fields) {
	fields := f
	if fields == nil {
		fields = Fields{}
	}

	l.lock()
	defer l.unlock()

	l.fields = fields
}

// Debugf formats and logs a debug message
func (l *logger) Debugf(format string, v ...interface{}) {
	// Debug is very verbose. Catch log-level early
	// to save unnecessary parsing
	if Dbg < l.logLevel {
		return
	}

	l.write(Dbg, fmt.Sprintf(format, v...))
}

// Debug logs a debug message
func (l *logger) Debug(v ...interface{}) {
	// Debug is very verbose. Catch log-level early
	// to save unnecessary parsing
	if Dbg < l.logLevel {
		return
	}

	l.write(Dbg, fmt.Sprint(v...))
}

// Infof formats and logs an informal message
func (l *logger) Infof(format string, v ...interface{}) {
	l.write(Nfo, fmt.Sprintf(format, v...))
}

// Info logs an informal message
func (l *logger) Info(v ...interface{}) {
	l.write(Nfo, fmt.Sprint(v...))
}

// Warningf formats and logs a warning message
func (l *logger) Warningf(format string, v ...interface{}) {
	l.write(Wrn, fmt.Sprintf(format, v...))
}

// Warning logs a warning message
func (l *logger) Warning(v ...interface{}) {
	l.write(Wrn, fmt.Sprint(v...))
}

// Errorf formats and logs an error message
func (l *logger) Errorf(format string, v ...interface{}) {
	l.write(Err, fmt.Sprintf(format, v...))
}

// Error logs an error message
func (l *logger) Error(v ...interface{}) {
	l.write(Err, fmt.Sprint(v...))
}

// Fatalf formats and logs an unrecoverable error message
func (l *logger) Fatalf(format string, v ...interface{}) {
	l.write(Ftl, fmt.Sprintf(format, v...))
	os.Exit(1)
}

// Fatal logs an unrecoverable error message
func (l *logger) Fatal(v ...interface{}) {
	l.write(Ftl, fmt.Sprint(v...))
	os.Exit(1)
}

// InstallHook installs a hook that will be called when a log event occurs
func (l *logger) InstallHook(logLevel LogLevel, hook HookFunc) {
	l.lock()
	defer l.unlock()

	if l.hooks == nil {
		l.hooks = make(map[LogLevel][]HookFunc)
	}

	l.hooks[logLevel] = append(l.hooks[logLevel], hook)
}

func (l *logger) write(logLevel LogLevel, msg string) {
	l.writeWithFields(logLevel, msg, l.GetFields(), l.GetFieldMapping())
}

func (l *logger) writeWithFields(logLevel LogLevel, msg string, fields Fields, fieldMapping FieldMapping) {
	now := time.Now()

	entryBuffer := bufferPool.Get().(*bytes.Buffer)
	defer bufferPool.Put(entryBuffer)
	entryBuffer.Reset()

	l.lock()
	defer l.unlock()

	if err := l.formatter.Format(entryBuffer, logLevel, msg, now, fields, fieldMapping); err != nil {
		fmt.Fprintf(os.Stderr, "error formatting the log entry: %v", err)
	}

	// Write to io.Writer if provided
	if l.writer != nil {
		if _, err := l.writer.Write(entryBuffer.Bytes()); err != nil {
			fmt.Fprintf(os.Stderr, "could not write log entry to io.Writer: %v", err)
		}
	}

	// Write to standard output if requested
	if l.stdOut {
		output := os.Stdout
		if logLevel > Wrn {
			output = os.Stderr
		}
		if _, err := entryBuffer.WriteTo(output); err != nil {
			fmt.Fprintf(os.Stderr, "could not write log entry to: %v", output)
		}
	}

	// Call any installed hooks
	if l.hooks != nil {
		for _, h := range l.hooks[logLevel] {
			h(now, logLevel, msg)
		}
	}
}

// SetFormatter sets or clears the writer of the logger
func (l *logger) SetFormatter(formatter Formatter) {
	l.lock()
	defer l.unlock()
	l.formatter = formatter
}

// GetFormatter gets the writer of the logger
func (l *logger) GetFormatter() Formatter {
	l.lock()
	defer l.unlock()
	return l.formatter
}

// SetWriter sets or clears the writer of the logger
func (l *logger) SetWriter(writer io.Writer) {
	l.lock()
	defer l.unlock()
	l.writer = writer
}

// SetLogLevel sets the log level of the logger
func (l *logger) SetLogLevel(logLevel LogLevel) {
	l.lock()
	defer l.unlock()
	l.logLevel = logLevel
}

// GetLogLevel implements Logger.GetLogLevel
func (l *logger) GetLogLevel() LogLevel {
	return l.logLevel
}

// GetFields implements Logger.GetFields
func (l *logger) GetFields() Fields {
	return l.fields
}

// GetFields implements Logger.GetFieldMapping
func (l *logger) GetFieldMapping() FieldMapping {
	return l.fieldMapping
}

// SetStdOut sets or clears writing to standard output
func (l *logger) SetStdOut(enable bool) {
	l.lock()
	defer l.unlock()
	l.stdOut = enable
}

// Debugf formats and logs a debug message
func Debugf(format string, v ...interface{}) {
	defaultLogger.Debugf(format, v...)
}

// Debug logs a debug message
func Debug(v ...interface{}) {
	defaultLogger.Debug(v...)
}

// Infof formats and logs an informal message
func Infof(format string, v ...interface{}) {
	defaultLogger.Infof(format, v...)
}

// Info logs an informal message
func Info(v ...interface{}) {
	defaultLogger.Info(v...)
}

// Warningf formats and logs a warning message
func Warningf(format string, v ...interface{}) {
	defaultLogger.Warningf(format, v...)
}

// Warning logs a warning message
func Warning(v ...interface{}) {
	defaultLogger.Warning(v...)
}

// Errorf formats and logs an error message
func Errorf(format string, v ...interface{}) {
	defaultLogger.Errorf(format, v...)
}

// Error logs an error message
func Error(v ...interface{}) {
	defaultLogger.Error(v...)
}

// Fatalf formats and logs an unrecoverable error message
func Fatalf(format string, v ...interface{}) {
	defaultLogger.Fatalf(format, v...)
}

// Fatal logs an unrecoverable error message
func Fatal(v ...interface{}) {
	defaultLogger.Fatal(v...)
}

// InstallHook installs a hook to the default logger
// that will be called when a log event occurs
func InstallHook(logLevel LogLevel, hook HookFunc) {
	defaultLogger.InstallHook(logLevel, hook)
}

// SetFieldMapping add custom field mapping for structured log
func SetFieldMapping(fieldMapping FieldMapping) {
	defaultLogger.SetFieldMapping(fieldMapping)
}

// SetFormatter sets the formatter to be used when outputting log entries
func SetFormatter(formatter Formatter) {
	defaultLogger.SetFormatter(formatter)
}

// SetWriter sets or clears the writer of the default logger
func SetWriter(writer io.Writer) {
	defaultLogger.SetWriter(writer)
}

// SetLogLevel sets the log level of the default logger
func SetLogLevel(logLevel LogLevel) {
	defaultLogger.SetLogLevel(logLevel)
}

// SetStdOut sets or clears writing to standard output of the default logger
func SetStdOut(enable bool) {
	defaultLogger.SetStdOut(enable)
}

// DefaultLogger returns the default logger
func DefaultLogger() MutableLogger {
	return defaultLogger
}

// SetFields sets the fields for the default logger. These fields will be appended to any
// child scope created with WithScope method.
func SetFields(fields Fields) {
	defaultLogger.SetFields(fields)
}

// SetGlobalFields sets the fields for the default logger. These fields will be appended to any
// Deprecated: Please use SetFields instead
func SetGlobalFields(fields Fields) {
	SetFields(fields)
}

// WithScope returns a new instance of Logger based on the default logger.
// Any fields from the default logger will be included to the new
// scoped Logger instance
func WithScope(fields Fields) Logger {
	return defaultLogger.WithScope(fields)
}
