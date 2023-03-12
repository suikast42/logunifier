package journald

import (
	"github.com/rs/zerolog"
	"github.com/suikast42/logunifier/internal/config"
	"github.com/suikast42/logunifier/pkg/model"
	"github.com/suikast42/logunifier/pkg/patterns"
	"github.com/suikast42/logunifier/pkg/utils"
	"os"
	"strings"
	"testing"
)

var (
	patternfactory *patterns.PatternFactory
	logger         zerolog.Logger
)

func init() {
	logger = config.Logger()
	_, err := patterns.Initialize()
	if err != nil {
		logger.Error().Err(err).Stack().Msg("Can't initialize pattern factory")
		os.Exit(1)
	}
	patternfactory = patterns.Instance()
}
func TestDefaultPatterContainerLogTest(t *testing.T) {
	expectedMessage := "TestMessage"
	log := TestMetaLogFromJournalD([]byte(testJournaldContainerLog), expectedMessage, t)

	parse := patternfactory.Parse(log)

	if parse.Message != "TestMessage" {
		t.Errorf("Expected %+v but got %+v", expectedMessage, parse.Message)
	}

}

func TestLogfmt(t *testing.T) {

	fmt, err := utils.DecodeLogFmt("esc=bar hasi=bongo xy=\"first line\nsecond line\"")
	if err != nil {
		t.Errorf("Unexepected error %s", err)
	}

	if len(fmt) != 3 {
		t.Errorf("Expected %+v but got %+v", 3, len(fmt))
	}
	for k, v := range fmt {
		logger.Info().Msgf("Key=%s Value=%s", k, v)
	}
}
func TestDockerServicelog(t *testing.T) {
	log := TestMetaLogFromJournalDFromConst([]byte(testJournaldDockerServiceLog), t)
	parsed := patternfactory.Parse(log)
	if parsed == nil {
		t.Error("Expected not nil but got nil")
	}

	grok, ok := parsed.Labels[(string(model.DynamicLabelUsedGrok))]

	if !ok {
		t.Error("Can't find pattern")
	}

	if grok != string(patterns.LogFmtPattern) {
		t.Errorf("Expected pattern %s but got %s", patterns.LogFmtPattern, grok)
	}
	if parsed.Log == nil {
		t.Error("Expected not nil but got nil")
	}

	if parsed.Log.Level != model.LogLevel_error {
		t.Errorf("Expected Log level %+v but got %+v", model.LogLevel_error, parsed.Log.Level)
	}

	if parsed.Log.LevelEmoji != model.LogLevelToEmoji(model.LogLevel_error) {
		t.Errorf("Expected Log level emoji  %+v but got %+v", model.LogLevelToEmoji(model.LogLevel_error), model.LogLevelToEmoji(parsed.Log.Level))
	}
	if !strings.Contains(parsed.Message, "collecting stats for") {
		t.Errorf("Expected Log message starts with  %+v but got %+v", "collecting stats for", parsed.Message)
	}

	if len(parsed.ProcessError.Reason) > 0 {
		t.Errorf("Expected no parse errors but got %+v", parsed.ProcessError)
	}
	//2023-03-10T18:53:52
	time := parsed.Timestamp.AsTime()
	year := 2023
	month := "March"
	day := 10
	hour := 18
	minute := 53
	second := 52
	if time.IsZero() {
		t.Errorf("Expected time not zero  but got %+v", time)
	}
	if time.Year() != year {
		t.Errorf("Expected %d  but got %d", year, time.Year())
	}
	if time.Month().String() != "March" {
		t.Errorf("Expected %s  but got %s", month, time.Month())
	}
	if time.Day() != day {
		t.Errorf("Expected %d  but got %d", day, time.Day())
	}
	if time.Minute() != minute {
		t.Errorf("Expected %d  but got %d", minute, time.Minute())
	}
	if time.Hour() != hour {
		t.Errorf("Expected %d  but got %d", hour, time.Hour())
	}
	if time.Second() != second {
		t.Errorf("Expected %d  but got %d", second, time.Second())
	}

}
