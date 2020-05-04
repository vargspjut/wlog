package wlog

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestJSONFormatter(t *testing.T) {
	type args struct {
		formatter Formatter
	}
	tests := []struct {
		name           string
		args           args
		resultTemplate string
	}{
		{
			"Parse to JSON",
			args{&JSONFormatter{}},
			`{"field1":"test value","level":"Info","msg":"test value","timestamp":"%s"}`},
		{
			"Parse to JSON with compact fields",
			args{&JSONFormatter{Compact: true}},
			`{"@l":"Info","@m":"test value","@t":"%s","field1":"test value"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now := time.Now()
			want := fmt.Sprintf(tt.resultTemplate, getTimestamp(now))

			logger := DefaultLogger()
			logger.SetLogLevel(Nfo)
			logger.SetFormatter(tt.args.formatter)
			logger.SetStdOut(false)
			logger.SetFields(Fields{"field1": "test value"})

			buf := &bytes.Buffer{}

			if err := logger.GetFormatter().Format(buf, Nfo, "test value", now, logger.GetFields(), logger.GetFieldMapping()); err != nil {
				t.Fatalf("failed to format the log entry, err: %s", err)
			}

			got := strings.TrimSuffix(buf.String(), "\n")

			if got != want {
				t.Errorf("formatter.Format() = %v, want %v", got, want)
			}
		})
	}
}
