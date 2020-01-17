package wlog

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"
)

func Test_JSONFormatter(t *testing.T) {

	now := time.Now()
	expected := fmt.Sprintf(`{"field1":"test value","level":"Info","msg":"test value","timestamp":"%s"}`, getTimestamp(now))

	SetLogLevel(Nfo)
	WithFields(Fields{"field1": "test value"})

	buf := &bytes.Buffer{}

	if err := logger.formatter.Format(buf, logger, "test value", now); err != nil {
		t.Fatalf("failed to format the log entry, err: %s", err)
	}

	result := strings.TrimSuffix(buf.String(), "\n")

	if result != expected {
		t.Fatalf("Expected: %s, got %s", expected, result)
	}
}
