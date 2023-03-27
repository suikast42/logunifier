package patterns

import (
	"testing"
	"time"
)

//2023-03-16T20:43:56.275171Z

func TestGrafanaTs(t *testing.T) {
	parse, err := time.Parse(time.RFC3339Nano, "2023-03-16T20:43:56.275171Z")
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	}

	if parse.IsZero() {
		t.Error("Expected no zero but got zero")
	}
}

func TestTsLevelMsgPattern(t *testing.T) {
	parse, err := time.Parse(time.RFC3339Nano, "2023-03-16T20:43:56.275171Z")
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	}

	if parse.IsZero() {
		t.Error("Expected no zero but got zero")
	}
}
