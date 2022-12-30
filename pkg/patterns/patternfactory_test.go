package patterns

import (
	"github.com/rs/zerolog"
	"github.com/suikast42/logunifier/internal/config"
	"github.com/suikast42/logunifier/pkg/model"
	"os"
	"reflect"
	"testing"
	"time"
)

var (
	patternfactory *PatternFactory
	logger         zerolog.Logger
)

func init() {
	logger = config.Logger()
	_, err := Initialize()
	if err != nil {
		logger.Error().Err(err).Stack().Msg("Can't initialize pattern factory")
		os.Exit(1)
	}
	patternfactory = Instance()
}
func TestParseTS_LEVEL_MSG(t *testing.T) {
	var log = "2022-12-08T12:21:02.594Z [ERROR] nomad.autopilot: 😜 Failed\nto reconcile current state with the desired state\nthird line mf\n1\n3"
	ts, _ := time.Parse(time.RFC3339, "2022-12-08T12:21:02.594Z")
	expected := ParseResult{
		LogLevel:  "ERROR",
		TimeStamp: ts,
		Msg:       "nomad.autopilot: 😜 Failed\nto reconcile current state with the desired state\nthird line mf\n1\n3",
	}
	parsed, err := patternfactory.Parse(TS_LEVEL_MSG, log)

	if err != nil {
		t.Error(err)
	}
	equal := reflect.DeepEqual(expected, parsed)
	if !equal {
		t.Errorf("Expected %+v but got %+v", expected, parsed)
	}
	//logger.Info().Msgf("%s", parsed)
}

func TestParseMSG_ONLY(t *testing.T) {
	var log = "sudo journalctl -f -u vector.service --since \"1 seconds ago\""
	expected := ParseResult{
		Msg: "sudo journalctl -f -u vector.service --since \"1 seconds ago\"",
	}
	parsed, err := patternfactory.Parse(MSG_ONLY, log)

	if err != nil {
		t.Error(err)
	}
	equal := reflect.DeepEqual(expected, parsed)
	if !equal {
		t.Errorf("Expected %+v but got %+v", expected, parsed)
	}
	//logger.Info().Msgf("%s", parsed)
}

func TestParseMSG_ONLY_With_Defaults(t *testing.T) {
	var log = "sudo journalctl -f -u vector.service --since \"1 seconds ago\""
	ts, _ := time.Parse(time.RFC3339, "2022-12-08T12:21:02.594Z")
	defaults := ParseResult{
		TimeStamp: ts,
		LogLevel:  "FakeLevel",
		Msg:       "That should be overwritten",
	}
	expected := ParseResult{
		Msg:       "sudo journalctl -f -u vector.service --since \"1 seconds ago\"",
		TimeStamp: ts,
		LogLevel:  "FakeLevel",
	}
	parsed, err := patternfactory.ParseWitDefaults(defaults, MSG_ONLY, log)

	if err != nil {
		t.Error(err)
	}
	equal := reflect.DeepEqual(expected, parsed)
	if !equal {
		t.Errorf("Expected %+v but got %+v", expected, parsed)
	}
	//logger.Info().Msgf("%s", parsed)
}

func TestEcsAggTags(t *testing.T) {
	entry := &model.EcsLogEntry{}
	if entry.GetTags() == nil {
		t.Error("Tags should be nil")
	}
	entry.Tags = append(entry.Tags, "1")
	entry.Tags = append(entry.Tags, "2")
	entry.Tags = append(entry.Tags, "3")
	if len(entry.Tags) != 3 {
		t.Errorf("Expect 3 elements bu got %d", len(entry.Tags))
	}
}
