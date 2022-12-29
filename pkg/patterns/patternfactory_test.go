package patterns

import (
	"github.com/rs/zerolog"
	"github.com/suikast42/logunifier/internal/config"
	"os"
	"reflect"
	"testing"
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
	expected := ParseResult{
		LogLevel:  "ERROR",
		TimeStamp: "2022-12-08T12:21:02.594Z",
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
