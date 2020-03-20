package wlog

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// Formatter is a base interface for output formatters, it has
// one method called Format which will be called when outputting
// the write entry
type Formatter interface {
	Format(w io.Writer, logLevel LogLevel, msg string, timestamp time.Time, fields Fields) error
}

// JSONFormatter used to output logs in JSON format
type JSONFormatter struct {
	Compact bool
}

func (j JSONFormatter) getKey(key string) string {
	if j.Compact {
		return "@" + key[:1]
	}
	return key
}

// Format implements Formatter.Format to support JSON
func (j JSONFormatter) Format(w io.Writer, logLevel LogLevel, msg string, timestamp time.Time, fields Fields) error {

	if fields == nil {
		fields = Fields{}
	}

	fields[j.getKey("msg")] = msg
	fields[j.getKey("timestamp")] = getTimestamp(timestamp)
	fields[j.getKey("level")] = logLevel.String()

	encoder := json.NewEncoder(w)
	if err := encoder.Encode(fields); err != nil {
		return fmt.Errorf("failed to marshal fields to JSON, %v", err)
	}

	return nil
}

// TextFormatter used to output logs in text format. This is the default
// formatter when creating a instance of wlog.
type TextFormatter struct{}

// Format Implements Formatter.Format to support Text
func (t TextFormatter) Format(w io.Writer, logLevel LogLevel, msg string, timestamp time.Time, fields Fields) error {

	// Write Date
	year, month, day := timestamp.Date()
	itoa(w, year, 4)
	writeString(w, "-")
	itoa(w, int(month), 2)
	writeString(w, "-")
	itoa(w, day, 2)

	writeString(w, " ")

	// Write time
	hour, min, sec := timestamp.Clock()
	itoa(w, hour, 2)
	writeString(w, ":")
	itoa(w, min, 2)
	writeString(w, ":")
	itoa(w, sec, 2)
	writeString(w, ":")
	itoa(w, timestamp.Nanosecond()/1e3, 6)

	writeString(w, " ")

	// Write log level
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

	writeString(w, level)

	// Append log message to buffer
	writeString(w, msg)

	if len(fields) > 0 {
		writeString(w, " [ ")
		writeFields(w, fields)
		writeString(w, "]")
	}

	if len(msg) == 0 || msg[len(msg)-1] != '\n' {
		writeString(w, "\n")
	}

	return nil
}

func writeFields(w io.Writer, fields Fields) {
	for key, value := range fields {
		writeString(w, key)
		writeString(w, "=")
		writeString(w, fmt.Sprintf("%v", value))
		writeString(w, ", ")
	}
}

func getTimestamp(now time.Time) string {
	year, month, day := now.Date()
	hour, min, sec := now.Clock()
	nano := now.Nanosecond() / 1e3

	format := "%d-%02d-%02d %02d:%02d:%02d:%06d"

	return fmt.Sprintf(format, year, month, day, hour, min, sec, nano)
}

func writeString(w io.Writer, str string) {
	if _, err := io.WriteString(w, str); err != nil {
		fmt.Fprintf(os.Stderr, "could not write entry log, err: %s", err)
	}
}

// Cheap integer to fixed-width decimal ASCII. Give a negative width to avoid zero-padding.
// NOTE: Taken from Go's std log package
func itoa(w io.Writer, i int, wid int) {
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

	if _, err := w.Write(b[bp:]); err != nil {
		fmt.Fprintf(os.Stderr, "failed adding zero padding to %d", i)
	}
}
