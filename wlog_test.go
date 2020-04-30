package wlog

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"
)

func TestHooks(t *testing.T) {

	// Install a hook to catch all NFO (Info) messages
	InstallHook(Nfo, func(timestamp time.Time, logLevel LogLevel, message string) {

		if logLevel != Nfo {
			t.Fatalf("Expected %s but got %s", Nfo, logLevel)
		}

		println("Hook: " + message)
	})

	// Install a hook to catch all WRN (Warning) messages
	InstallHook(Wrn, func(timestamp time.Time, logLevel LogLevel, message string) {

		if logLevel != Wrn {
			t.Fatalf("Expected %s but got %s", Wrn, logLevel)
		}

		println("Hook: " + message)
	})

	Info("This is a NFO log entry")
	Warning("This is a WRN log entry")
	Error("This is a ERR log entry. No hooks installed for this level")
}

func TestLogging(t *testing.T) {

	SetFormatter(&JSONFormatter{
		FieldMapping: FieldMapping{"tenantId": "tid"},
		Compact:      true,
	})

	SetStdOut(false)

	w := &bytes.Buffer{}

	SetWriter(w)

	AddFieldMapping(FieldMapping{"name": "n", "address": "addr"})

	logger := WithScope(Fields{"tenantId": "1223456", "name": "user", "address": "my street"})

	logger.Info("This is a test")

	reader := bytes.NewReader(w.Bytes())
	d := json.NewDecoder(reader)

	var data map[string]string
	d.Decode(&data)

	expectedKeys := []string{"@t", "@m", "@l", "tid", "n", "addr"}

	for _, k := range expectedKeys {
		if _, ok := data[k]; !ok {
			t.Fatalf("should contain a key: %s", k)
		}
	}

}