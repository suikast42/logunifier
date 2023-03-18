package model

import (
	"github.com/rs/zerolog"
	"github.com/suikast42/logunifier/internal/config"
	"testing"
)

var (
	logger zerolog.Logger
)

func init() {
	logger = config.Logger()
}

func TestEcsAggTags(t *testing.T) {
	entry := &EcsLogEntry{}

	entry.Tags = append(entry.Tags, "1")
	entry.Tags = append(entry.Tags, "2")
	entry.Tags = append(entry.Tags, "3")
	if len(entry.Tags) != 3 {
		t.Errorf("Expect 3 elements bu got %d", len(entry.Tags))
	}
}

func TestEcsAggLabels(t *testing.T) {
	entry := NewEcsLogEntry()
	entry.Labels["1"] = "1"
	entry.Labels["2"] = "2"
	entry.Labels["3"] = "3"
	entry.Labels["4"] = "4"
	if len(entry.Labels) != 4 {
		t.Errorf("Expect 4 elements bu got %d", len(entry.Labels))
	}
}
