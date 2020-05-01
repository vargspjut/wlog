package wlog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// Formatter is a base interface for output formatters, it has
// one method called Format which will be called when outputting
// the write entry
type Formatter interface {
	Format(w io.Writer, logLevel LogLevel, msg string, timestamp time.Time, fields Fields) error
	SetFieldMapping(fieldMapping FieldMapping)
}

// JSONFormatter used to output logs in JSON format
type JSONFormatter struct {
	FieldMapping FieldMapping
	Compact      bool
}

func (j JSONFormatter) getKey(key string) string {
	if j.Compact {
		var mappedKey string
		var ok bool

		if mappedKey, ok = j.FieldMapping[key]; !ok {
			mappedKey = key
		}

		return mappedKey
	}
	return key
}

// SetFieldMapping add custom field mapping for structured log
func (j *JSONFormatter) SetFieldMapping(fieldMapping FieldMapping) {
	var mapping = make(FieldMapping, 3)

	// first add the custom mappings
	for k, v := range fieldMapping {
		if strings.HasPrefix(v, "@") {
			fmt.Fprintf(os.Stderr, "value cannot be prefixed with @: %s", v)
		}
		mapping[k] = v
	}

	// add the default mappings
	for k, v := range j.FieldMapping {
		mapping[k] = v
	}

	j.FieldMapping = mapping
}

// Format implements Formatter.Format to support JSON
func (j JSONFormatter) Format(w io.Writer, logLevel LogLevel, msg string, timestamp time.Time, fields Fields) error {

	// Standard fields
	out := Fields{
		"@m": msg,
		"@t": getTimestamp(timestamp),
		"@l": logLevel.String(),
	}

	// And any custom ones
	for k, v := range fields {
		out[j.getKey(k)] = v
	}

	encoder := json.NewEncoder(w)
	if err := encoder.Encode(out); err != nil {
		return fmt.Errorf("failed to marshal fields to JSON, %v", err)
	}

	return nil
}

// TextFormatter used to output logs in text format. This is the default
// formatter when creating a instance of wlog.
type TextFormatter struct{}

// SetFieldMapping add custom field mapping for structured log
func (j TextFormatter) SetFieldMapping(fieldMapping FieldMapping) {
	fmt.Fprintf(os.Stderr, "mapping fields is only supported by JSONFormatter")
}

// Format Implements Formatter.Format to support Text
func (t TextFormatter) Format(w io.Writer, logLevel LogLevel, msg string, timestamp time.Time, fields Fields) error {

	// Write date and time
	writeString(w, getTimestamp(timestamp))

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
		writeString(w, " [")
		writeFields(w, fields)
		writeString(w, "]")
	}

	if len(msg) == 0 || msg[len(msg)-1] != '\n' {
		writeString(w, "\n")
	}

	return nil
}

func writeFields(w io.Writer, fields Fields) {
	count := len(fields)
	idx := 0
	for key, value := range fields {
		idx++
		writeString(w, key)
		writeString(w, ": ")
		writeString(w, fmt.Sprintf("%v", value))
		if idx < count {
			writeString(w, ", ")
		}
	}
}

func getTimestamp(timestamp time.Time) string {

	w := bufferPool.Get().(*bytes.Buffer)
	defer bufferPool.Put(w)
	w.Reset()

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

	return string(w.Bytes())
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
