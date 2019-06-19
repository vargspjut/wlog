package wlog

import (
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
