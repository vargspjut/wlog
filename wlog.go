package wlog

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Config allows configuration of the wlogger
type Config struct {
	LogLevel        LogLevel
	Path            string
	TruncateOnStart bool
	StdOut          bool
	Formatter       Formatter
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
	logger *Logger
)

func init() {
	logger = New(nil, Nfo, true)
}

// New initialized a new logger object
func New(writer io.Writer, logLevel LogLevel, stdOut bool) *Logger {
	return &Logger{
		writer:    writer,
		logLevel:  logLevel,
		stdOut:    stdOut,
		formatter: TextFormatter{},
		fields:    Fields{},
	}
}

// WLogger is the interface that wlog loggers implements
type WLogger interface {
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
	GetLogLevel() LogLevel
}

// Fields is a map containing the fields that will be added to every log entry
type Fields map[string]interface{}

// Logger provides logging levels to standard logger library
type Logger struct {
	writer    io.Writer
	logLevel  LogLevel
	stdOut    bool
	lock      sync.Mutex
	hooks     map[LogLevel][]HookFunc
	fields    Fields
	formatter Formatter
}

var bufferPool = sync.Pool{New: func() interface{} {
	return new(bytes.Buffer)
}}

// Configure configures the logger
func (l *Logger) Configure(cfg *Config) {
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
		l.SetWriter(nil)
	}
}

// WithFields method attach fields to the logger global scope
func (l *Logger) WithFields(fields Fields) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.formatter = JSONFormatter{}
	l.fields = fields
}

// WithScope returns a new instance of ScopedLogger and its fields property
// will contain the fields added to the global instance of Logger
func (l *Logger) WithScope(fields Fields) *ScopedLogger {
	scopeFields := fields
	//copy the parent scope fields
	for k, v := range l.fields {
		scopeFields[k] = v
	}
	return &ScopedLogger{scopeFields, l, JSONFormatter{}}
}

// SetGlobalFields set fields in a global wlog instance. These fields will be appended to any
// child scope created with wlog.WithScope method.
func (l *Logger) SetGlobalFields(f Fields) {
	fields := f
	if fields == nil {
		fields = Fields{}
	}

	l.lock.Lock()
	defer l.lock.Unlock()
	l.fields = fields
	l.formatter = JSONFormatter{}
}

// Debugf formats and logs a debug message
func (l *Logger) Debugf(format string, v ...interface{}) {
	// Debug is very verbose. Catch log-level early
	// to save unnecessary parsing
	if Dbg < l.logLevel {
		return
	}

	l.writef(Dbg, fmt.Sprintf(format, v...))
}

// Debug logs a debug message
func (l *Logger) Debug(v ...interface{}) {
	// Debug is very verbose. Catch log-level early
	// to save unnecessary parsing
	if Dbg < l.logLevel {
		return
	}

	l.writef(Dbg, fmt.Sprint(v...))
}

// Infof formats and logs an informal message
func (l *Logger) Infof(format string, v ...interface{}) {
	l.writef(Nfo, fmt.Sprintf(format, v...))
}

// Info logs an informal message
func (l *Logger) Info(v ...interface{}) {
	l.writef(Nfo, fmt.Sprint(v...))
}

// Warningf formats and logs a warning message
func (l *Logger) Warningf(format string, v ...interface{}) {
	l.writef(Wrn, fmt.Sprintf(format, v...))
}

// Warning logs a warning message
func (l *Logger) Warning(v ...interface{}) {
	l.writef(Wrn, fmt.Sprint(v...))
}

// Errorf formats and logs an error message
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.writef(Err, fmt.Sprintf(format, v...))
}

// Error logs an error message
func (l *Logger) Error(v ...interface{}) {
	l.writef(Err, fmt.Sprint(v...))
}

// Fatalf formats and logs an unrecoverable error message
func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.writef(Ftl, fmt.Sprintf(format, v...))
	os.Exit(1)
}

// Fatal logs an unrecoverable error message
func (l *Logger) Fatal(v ...interface{}) {
	l.writef(Ftl, fmt.Sprint(v...))
	os.Exit(1)
}

// InstallHook installs a hook that will be called when a log event occurs
func (l *Logger) InstallHook(logLevel LogLevel, hook HookFunc) {
	l.lock.Lock()
	defer l.lock.Unlock()

	if l.hooks == nil {
		l.hooks = make(map[LogLevel][]HookFunc)
	}

	l.hooks[logLevel] = append(l.hooks[logLevel], hook)
}

func (l *Logger) writef(logLevel LogLevel, msg string) {
	now := time.Now()

	entryBuffer := bufferPool.Get().(*bytes.Buffer)
	defer bufferPool.Put(entryBuffer)
	entryBuffer.Reset()

	if err := l.formatter.Format(entryBuffer, logLevel, l, msg, now); err != nil {
		fmt.Fprintf(os.Stderr, "error formatting the log entry: %v", err)
	}

	write(l, entryBuffer, logLevel, msg, now)
}

// write writes a log entry to file and possibly to standard output
func write(l *Logger, buffer *bytes.Buffer, logLevel LogLevel, msg string, entryTime time.Time) {

	logEntry := buffer.Bytes()

	// Write to file if provided
	if l.writer != nil {
		if _, err := l.writer.Write(logEntry); err != nil {
			fmt.Fprintf(os.Stderr, "could not write log entry to the file, err: %v", err)
		}
	}

	// Write to standard output if requested
	if l.stdOut {
		output := os.Stdout
		if logLevel > Wrn {
			output = os.Stderr
		}
		if _, err := buffer.WriteTo(output); err != nil {
			fmt.Fprintf(os.Stderr, "could not write log entry to: %v", output)
		}
	}

	// Call any installed hooks
	if l.hooks != nil {
		for _, h := range l.hooks[logLevel] {
			h(entryTime, logLevel, msg)
		}
	}
}

// SetFormatter sets or clears the writer of the logger
func (l *Logger) SetFormatter(formatter Formatter) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.formatter = formatter
}

// SetWriter sets or clears the writer of the logger
func (l *Logger) SetWriter(writer io.Writer) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.writer = writer
}

// SetLogLevel sets the log level of the logger
func (l *Logger) SetLogLevel(logLevel LogLevel) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.logLevel = logLevel
}

// GetLogLevel implements WLogger.GetLogLevel
func (l *Logger) GetLogLevel() LogLevel {
	return l.logLevel
}

// GetFields implements WLogger.GetFields
func (l *Logger) GetFields() Fields {
	return l.fields
}

// SetStdOut sets or clears writing to standard output
func (l *Logger) SetStdOut(enable bool) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.stdOut = enable
}

// Debugf formats and logs a debug message
func Debugf(format string, v ...interface{}) {
	logger.Debugf(format, v...)
}

// Debug logs a debug message
func Debug(v ...interface{}) {
	logger.Debug(v...)
}

// Infof formats and logs an informal message
func Infof(format string, v ...interface{}) {
	logger.Infof(format, v...)
}

// Info logs an informal message
func Info(v ...interface{}) {
	logger.Info(v...)
}

// Warningf formats and logs a warning message
func Warningf(format string, v ...interface{}) {
	logger.Warningf(format, v...)
}

// Warning logs a warning message
func Warning(v ...interface{}) {
	logger.Warning(v...)
}

// Errorf formats and logs an error message
func Errorf(format string, v ...interface{}) {
	logger.Errorf(format, v...)
}

// Error logs an error message
func Error(v ...interface{}) {
	logger.Error(v...)
}

// Fatalf formats and logs an unrecoverable error message
func Fatalf(format string, v ...interface{}) {
	logger.Fatalf(format, v...)
}

// Fatal logs an unrecoverable error message
func Fatal(v ...interface{}) {
	logger.Fatal(v...)
}

// InstallHook installs a hook to the default logger
// that will be called when a log event occurs
func InstallHook(logLevel LogLevel, hook HookFunc) {
	logger.InstallHook(logLevel, hook)
}

// SetFormatter sets the formatter to be used when outputting log entries
func SetFormatter(formatter Formatter) {
	logger.SetFormatter(formatter)
}

// SetWriter sets or clears the writer of the default logger
func SetWriter(writer io.Writer) {
	logger.SetWriter(writer)
}

// SetLogLevel sets the log level of the default logger
func SetLogLevel(logLevel LogLevel) {
	logger.SetLogLevel(logLevel)
}

// SetStdOut sets or clears writing to standard output of the default logger
func SetStdOut(enable bool) {
	logger.SetStdOut(enable)
}

// DefaultLogger returns the default logger
func DefaultLogger() *Logger {
	return logger
}

// SetGlobalFields include Fields to the Logger instance setting the JSONFormatter by default
func SetGlobalFields(fields Fields) {
	logger.SetGlobalFields(fields)
}

// WithScope returns a new instance of ScopedLogger and its fields property
// will contain the fields added to the global instance of Logger
func WithScope(fields Fields) *ScopedLogger {
	return logger.WithScope(fields)
}

// WithFields method attach fields to the logger global scope
func WithFields(fields Fields) {
	logger.WithFields(fields)
}