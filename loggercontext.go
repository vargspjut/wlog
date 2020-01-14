package wlog

import (
	"fmt"
	"os"
	"time"
)

// Fields are the fields that will be included in a JSON log output
type Fields map[string]interface{}

// LoggerContext represents the logger current context. This is not meant to be used
// by its own, you can get a instead of LoggerContext using the method WithContext
// and passing the fields that should be included in the context.
//
// wlog.WithContext(Fields{"firstname": "John", "lastname": "Smith"})
type LoggerContext struct {
	logger    *Logger
	fields    Fields
	formatter Formatter
}

// WithContext get a list of Fields and create a new LoggerContext instance.
func (c *LoggerContext) WithFields(f Fields) *LoggerContext {
	newContextFields := make(Fields, len(c.fields)+len(f))

	// first the fields on the previous context are added
	for k, v := range c.fields {
		newContextFields[k] = v
	}
	// add the new fields
	for k, v := range f {
		newContextFields[k] = v
	}
	return &LoggerContext{logger: c.logger, fields: newContextFields, formatter: c.formatter}
}

func getTimestamp(now time.Time) string {
	year, month, day := now.Date()
	hour, min, sec := now.Clock()
	nano := now.Nanosecond() / 1e3

	format := "%d-%02d-%02d %02d:%02d:%02d:%06d"

	return fmt.Sprintf(format, year, month, day, hour, min, sec, nano)
}

func (c *LoggerContext) write(logLevel LogLevel, msg string) {
	if logLevel < c.logger.logLevel {
		return
	}

	now := time.Now()

	switch c.formatter.(type) {
	case JSONFormatter:
		c.fields["msg"] = msg
		c.fields["level"] = logLevel.String()
		c.fields["timestamp"] = getTimestamp(now)

		c.logger.lock.Lock()
		defer c.logger.lock.Unlock()

		c.logger.buffer = c.logger.buffer[:0]

		formattedFields, err := c.formatter.Format(c)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to format context fields, %v\n", err)
			return
		}

		c.logger.buffer = append(c.logger.buffer, formattedFields...)

	default:
		c.logger.write(logLevel, msg)
		return
	}

	// Write to file if provided
	if c.logger.writer != nil {
		c.logger.writer.Write(c.logger.buffer)
	}

	// Write to standard output if requested
	if c.logger.stdOut {
		if logLevel > Wrn {
			os.Stderr.Write(c.logger.buffer)
		} else {
			os.Stdout.Write(c.logger.buffer)
		}
	}

	// Call any installed hooks
	if c.logger.hooks != nil {
		for _, h := range c.logger.hooks[logLevel] {
			h(now, logLevel, msg)
		}
	}

}

// Info emits an INFO log entry
func (c *LoggerContext) Info(v ...interface{}) {
	c.write(Nfo, fmt.Sprint(v...))
}

// Infof emits an INFO log entry
func (c *LoggerContext) Infof(format string, v ...interface{}) {
	c.write(Nfo, fmt.Sprintf(format, v...))
}

// Debug emits a DEBUG log entry
func (c *LoggerContext) Debug(v ...interface{}) {
	if Dbg < c.logger.logLevel {
		return
	}
	c.write(Dbg, fmt.Sprint(v...))
}

// Debugf emits a DEBUG log entry
func (c *LoggerContext) Debugf(format string, v ...interface{}) {
	if Dbg < c.logger.logLevel {
		return
	}
	c.write(Dbg, fmt.Sprintf(format, v...))
}

// Error emits an ERROR log entry
func (c *LoggerContext) Error(v ...interface{}) {
	c.write(Err, fmt.Sprint(v...))
}

// Errorf emits an ERROR log entry
func (c *LoggerContext) Errorf(format string, v ...interface{}) {
	c.write(Err, fmt.Sprintf(format, v...))
}

// Warning emits a WARNING log entry
func (c *LoggerContext) Warning(v ...interface{}) {
	c.write(Wrn, fmt.Sprint(v...))
}

// Warningf emits a WARNING log entry
func (c *LoggerContext) Warningf(format string, v ...interface{}) {
	c.write(Wrn, fmt.Sprintf(format, v...))
}

// Fatal emits a FATAL log entry
func (c *LoggerContext) Fatal(v ...interface{}) {
	c.write(Ftl, fmt.Sprint(v...))
	os.Exit(1)
}

// Fatalf emits a FATAL log entry
func (c *LoggerContext) Fatalf(format string, v ...interface{}) {
	c.write(Ftl, fmt.Sprintf(format, v...))
	os.Exit(1)
}
