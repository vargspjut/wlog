package wlog

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// Formatter is a base interface for output formatters, it has
// one method called Format which will be called when outputting
// the write entry
type Formatter interface {
	Format(c *LoggerContext) ([]byte, error)
}

// JSONFormatter used to output logs in JSON format
type JSONFormatter struct{}

// Implements Formatter.Format
func (j JSONFormatter) Format(c *LoggerContext) ([]byte, error) {

	b := &bytes.Buffer{}
	encoder := json.NewEncoder(b)

	if err := encoder.Encode(c.fields); err != nil {
		return nil, fmt.Errorf("failed to marshal fields to JSON, %v", err)
	}

	return b.Bytes(), nil

}
