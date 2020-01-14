package wlog

import (
	"fmt"
	"testing"
)

func Test_JSONFormatter(t *testing.T) {
	expected := "{\"field1\":\"test value\"}\n"

	logger := WithFields(Fields{"field1": "test value"})
	result, _ := logger.formatter.Format(logger)
	strResult := fmt.Sprintf("%s", result)

	if strResult != expected {
		t.Fatalf("Expected: %s, got %s", expected, strResult)
	}
}
