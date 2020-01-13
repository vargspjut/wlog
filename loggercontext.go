package wlog

import (
	"fmt"
	"os"
)

type Fields map[string]interface{}

// LoggerContext represents the logger current context. This is not meant to be used
// by its own, you can get a instead of LoggerContext using the method WithContext
// and passing the fields that should be included in the context.
//
// wlog.WithContext(Fields{"firstname": "John", "lastname": "Smith"})
type LoggerContext struct {
	logger *Logger
	fields Fields
}

// WithContext get a list of Fields and create a new LoggerContext instance.
func (ctx *LoggerContext) WithContext(f Fields) *LoggerContext {
	newContextFields := make(Fields, len(ctx.fields)+len(f))

	// first the fields on the previous context are added
	for k, v := range ctx.fields {
		newContextFields[k] = v
	}
	// add the new fields
	for k, v := range f {
		newContextFields[k] = v
	}
	return &LoggerContext{logger: ctx.logger, fields: newContextFields}
}

func (ctx *LoggerContext) write(logLevel LogLevel, msg string) {

	if logLevel < ctx.logger.logLevel {
		return
	}

	switch ctx.logger.formatter.(type) {
	case JSONFormatter:
		updatedContext := ctx.WithContext(Fields{"msg": msg})
		formattedFields, err := ctx.logger.formatter.Format(updatedContext)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to format context fields, %v\n", err)
			return
		}
		ctx.logger.Write(logLevel, formattedFields)
	default:
		ctx.logger.Write(logLevel, msg)
	}
}

// Emits an INFO log entry
func (ctx *LoggerContext) Info(v ...interface{}) {
	ctx.write(Nfo, fmt.Sprint(v...))
}

// Emits an INFO log entry
func (ctx *LoggerContext) Infof(format string, v ...interface{}) {
	ctx.write(Nfo, fmt.Sprintf(format, v...))
}

// Emits a DEBUG log entry
func (ctx *LoggerContext) Debug(v ...interface{}) {
	if Dbg < ctx.logger.logLevel {
		return
	}
	ctx.write(Dbg, fmt.Sprint(v...))
}

// Emits a DEBUG log entry
func (ctx *LoggerContext) Debugf(format string, v ...interface{}) {
	if Dbg < ctx.logger.logLevel {
		return
	}
	ctx.write(Dbg, fmt.Sprintf(format, v...))
}

// Emits an ERROR log entry
func (ctx *LoggerContext) Error(v ...interface{}) {
	ctx.write(Err, fmt.Sprint(v...))
}

// Emits an ERROR log entry
func (ctx *LoggerContext) Errorf(format string, v ...interface{}) {
	ctx.write(Err, fmt.Sprintf(format, v...))
}

// Emits a WARNING log entry
func (ctx *LoggerContext) Warning(v ...interface{}) {
	ctx.write(Wrn, fmt.Sprint(v...))
}

// Emits a WARNING log entry
func (ctx *LoggerContext) Warningf(format string, v ...interface{}) {
	ctx.write(Wrn, fmt.Sprintf(format, v...))
}

// Emits a FATAL log entry
func (ctx *LoggerContext) Fatal(v ...interface{}) {
	ctx.write(Ftl, fmt.Sprint(v...))
	os.Exit(1)
}

// Emits a FATAL log entry
func (ctx *LoggerContext) Fatalf(format string, v ...interface{}) {
	ctx.write(Ftl, fmt.Sprintf(format, v...))
	os.Exit(1)
}
