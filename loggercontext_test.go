package wlog

import (
	"testing"
	"time"
)

func Test_LoggerContext_Hooks(t *testing.T) {
	// Install a hook to catch all NFO (Info) messages
	InstallHook(Nfo, func(timestamp time.Time, logLevel LogLevel, message string) {
		if logLevel != Nfo {
			t.Fatalf("Expected %s but got %s", Nfo, logLevel)
		}
		println("Hook: " + message)
	})

	loggerContext := WithFields(Fields{"test": "some value"})
	loggerContext.Info("This is a INFO log entry")
}

func Test_WithFields_NilValue(t *testing.T) {

	logger := WithFields(nil)
	if logger.fields == nil {
		t.Fatalf("Expected logger.fields to not be nil.")
	}
}

func Test_WithFields_DefaultProps(t *testing.T) {
	expected := 4
	defaultFields := []string{"timestamp", "level", "msg"}

	logger := WithFields(Fields{"field1": 10})
	logger.Info("This is a INFO log entry")

	totalFields := len(logger.fields)

	if  totalFields != expected {
		t.Fatalf("Expected %d, got %d", expected, totalFields)
	}

	for _, field := range defaultFields {
		if _, found := logger.fields[field]; !found {
			t.Fatalf("Expected the default field \"%s\" is missing.", field)
		}
	}
}

func Test_WithFields(t *testing.T) {
	type args struct {
		fields Fields
		nested bool
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			"LoggerContext should contain one field",
			args{Fields{"field1": 10}, false},
			1,
		},
		{
			"LoggerContext should contain two fields",
			args{Fields{"field1": 10, "field2": 20}, false},
			2,
		},
		{
			"Nested LoggerContext should get fields from parent",
			args{Fields{"field1": 10}, true},
			2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.args.nested {
				logger := WithFields(Fields{"parentField": "some value"})
				if got := logger.WithFields(tt.args.fields); len(got.fields) != tt.want {
					t.Errorf("WithFields() = %v, want %v", len(got.fields), tt.want)
				}
				return
			}
			if got := WithFields(tt.args.fields); len(got.fields) != tt.want {
				t.Errorf("WithFields() = %v, want %v", len(got.fields), tt.want)
			}
		})
	}
}
