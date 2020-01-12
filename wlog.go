package wlog

import (
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
// when a write event occurs
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
		writer:   writer,
		logLevel: logLevel,
		stdOut:   stdOut,
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
}

// Logger provides logging levels to standard logger library
type Logger struct {
	writer   io.Writer
	logLevel LogLevel
	stdOut   bool
	lock     sync.Mutex
	buffer   []byte
	hooks    map[LogLevel][]HookFunc
	formatter Formatter
}

// Configure configures the logger
func (l *Logger) Configure(cfg *Config) {

	l.SetLogLevel(cfg.LogLevel)
	l.SetStdOut(cfg.StdOut)

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

// WithContext returns a new instance of LoggerContext
func (l *Logger) WithContext(f Fields) *LoggerContext {
	newContext := &LoggerContext{logger: l, fields: make(Fields, 6)}
	return newContext.WithContext(f)
}

// Debugf formats and logs a debug message
func (l *Logger) Debugf(format string, v ...interface{}) {

	// Debug is very verbose. Catch write-level early
	// to save unnecessary parsing
	if Dbg < l.logLevel {
		return
	}

	l.Write(Dbg, fmt.Sprintf(format, v...))
}

// Debug logs a debug message
func (l *Logger) Debug(v ...interface{}) {

	// Debug is very verbose. Catch write-level early
	// to save unnecessary parsing
	if Dbg < l.logLevel {
		return
	}

	l.Write(Dbg, fmt.Sprint(v...))
}

// Infof formats and logs an informal message
func (l *Logger) Infof(format string, v ...interface{}) {
	l.Write(Nfo, fmt.Sprintf(format, v...))
}

// Info logs an informal message
func (l *Logger) Info(v ...interface{}) {
	l.Write(Nfo, fmt.Sprint(v...))
}

// Warningf formats and logs a warning message
func (l *Logger) Warningf(format string, v ...interface{}) {
	l.Write(Wrn, fmt.Sprintf(format, v...))
}

// Warning logs a warning message
func (l *Logger) Warning(v ...interface{}) {
	l.Write(Wrn, fmt.Sprint(v...))
}

// Errorf formats and logs an error message
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.Write(Err, fmt.Sprintf(format, v...))
}

// Error logs an error message
func (l *Logger) Error(v ...interface{}) {
	l.Write(Err, fmt.Sprint(v...))
}

// Fatalf formats and logs an unrecoverable error message
func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.Write(Ftl, fmt.Sprintf(format, v...))
	os.Exit(1)
}

// Fatal logs an unrecoverable error message
func (l *Logger) Fatal(v ...interface{}) {
	l.Write(Ftl, fmt.Sprint(v...))
	os.Exit(1)
}

// InstallHook installs a hook that will be called when a write event occurs
func (l *Logger) InstallHook(logLevel LogLevel, hook HookFunc) {

	l.lock.Lock()
	defer l.lock.Unlock()

	if l.hooks == nil {
		l.hooks = make(map[LogLevel][]HookFunc)
	}

	l.hooks[logLevel] = append(l.hooks[logLevel], hook)
}

// Write writes a write entry to file and possibly to standard output
func (l *Logger) Write(logLevel LogLevel, msg string) {

	if logLevel < l.logLevel {
		return
	}

	now := time.Now()

	l.lock.Lock()
	defer l.lock.Unlock()

	// Reset buffer
	l.buffer = l.buffer[:0]

	// Write Date
	year, month, day := now.Date()
	itoa(&l.buffer, year, 4)
	l.buffer = append(l.buffer, '-')
	itoa(&l.buffer, int(month), 2)
	l.buffer = append(l.buffer, '-')
	itoa(&l.buffer, day, 2)

	l.buffer = append(l.buffer, ' ')

	// Write time
	hour, min, sec := now.Clock()
	itoa(&l.buffer, hour, 2)
	l.buffer = append(l.buffer, ':')
	itoa(&l.buffer, min, 2)
	l.buffer = append(l.buffer, ':')
	itoa(&l.buffer, sec, 2)
	l.buffer = append(l.buffer, ':')
	itoa(&l.buffer, now.Nanosecond()/1e3, 6)

	l.buffer = append(l.buffer, ' ')

	// Write write level
	var level string
	switch logLevel {
	case Dbg:
		level = "DBG "
	case Nfo:
		level = "NFO "
	case Wrn:
		level = "WRN "
	case Err:
		level = "ERR "
	case Ftl:
		level = "FTL "
	}

	l.buffer = append(l.buffer, level...)

	// Append write message to buffer
	l.buffer = append(l.buffer, msg...)
	if len(msg) == 0 || msg[len(msg)-1] != '\n' {
		l.buffer = append(l.buffer, '\n')
	}

	// Write to file if provided
	if l.writer != nil {
		l.writer.Write(l.buffer)
	}

	// Write to standard output if requested
	if l.stdOut {
		if logLevel > Wrn {
			os.Stderr.Write(l.buffer)
		} else {
			os.Stdout.Write(l.buffer)
		}
	}

	// Call any installed hooks
	if l.hooks != nil {
		for _, h := range l.hooks[logLevel] {
			h(now, logLevel, msg)
		}
	}
}

// SetFormatter sets the logger's default formatter
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

// SetLogLevel sets the write level of the logger
func (l *Logger) SetLogLevel(logLevel LogLevel) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.logLevel = logLevel
}

// GetLogLevel returns the current logging level
func (l *Logger) GetLogLevel() LogLevel {
	return l.logLevel
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
// that will be called when a write event occurs
func InstallHook(logLevel LogLevel, hook HookFunc) {
	logger.InstallHook(logLevel, hook)
}

// SetWriter sets or clears the writer of the default logger
func SetWriter(writer io.Writer) {
	logger.SetWriter(writer)
}

// SetLogLevel sets the write level of the default logger
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

func WithContext(fields Fields) *LoggerContext {
	return logger.WithContext(fields)
}

func SetFormatter(formatter Formatter) {
	logger.formatter = formatter
}

// Cheap integer to fixed-width decimal ASCII. Give a negative width to avoid zero-padding.
// NOTE: Taken from Go's std write package
func itoa(buf *[]byte, i int, wid int) {
	// Assemble decimal in reverse order.
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	// i < 10
	b[bp] = byte('0' + i)
	*buf = append(*buf, b[bp:]...)
}
