package journald

import (
	"github.com/rs/zerolog"
	"github.com/suikast42/logunifier/internal/config"
	"github.com/suikast42/logunifier/pkg/model"
	"github.com/suikast42/logunifier/pkg/patterns"
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

func TestDockerServiceLog(t *testing.T) {
	log := TestMetaLogFromJournalDFromConst([]byte(testJournaldDockerServiceLog), t)
	parsed := patternfactory.Parse(log)
	if parsed == nil {
		t.Error("Expected not nil but got nil")
	}

	grok, ok := parsed.Labels[(string(model.DynamicLabelUsedGrok))]

	if !ok {
		t.Error("Can't find pattern")
	}

	if grok != model.MetaLog_LogFmt.String() {
		t.Errorf("Expected pattern %s but got %s", model.MetaLog_LogFmt.String(), grok)
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

func TestGrafanaLog(t *testing.T) {
	log := TestMetaLogFromJournalDFromConst([]byte(testJournaldGrafanaLog), t)
	parsed := patternfactory.Parse(log)
	if parsed == nil {
		t.Error("Expected not nil but got nil")
	}

	grok, ok := parsed.Labels[(string(model.DynamicLabelUsedGrok))]

	if !ok {
		t.Error("Can't find pattern")
	}

	if grok != model.MetaLog_LogFmt.String() {
		t.Errorf("Expected pattern [%s] but got [%s]", model.MetaLog_LogFmt.String(), grok)
	}
	if parsed.Log == nil {
		t.Error("Expected not nil but got nil")
	}

	if parsed.Log.Level != model.LogLevel_info {
		t.Errorf("Expected Log level %+v but got %+v", model.LogLevel_info, parsed.Log.Level)
	}

	if parsed.Log.LevelEmoji != model.LogLevelToEmoji(model.LogLevel_info) {
		t.Errorf("Expected Log level emoji  %+v but got %+v", model.LogLevelToEmoji(model.LogLevel_info), model.LogLevelToEmoji(parsed.Log.Level))
	}
	if !strings.Contains(parsed.Message, "Initialized channel handler") {
		t.Errorf("Expected Log message starts with  %+v but got %+v", "Initialized channel handler", parsed.Message)
	}

	if len(parsed.ProcessError.Reason) > 0 {
		t.Errorf("Expected no parse errors but got %+v", parsed.ProcessError)
	}
	//2023-03-16T20:43:56.274825539Z
	time := parsed.Timestamp.AsTime()
	year := 2023
	month := "March"
	day := 16
	hour := 20
	minute := 43
	second := 56
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

func TestTraefikInvalidLogFmt(t *testing.T) {
	log := TestMetaLogFromJournalDFromConst([]byte(testJournaldTraefikInvalidLogFmt), t)
	parsed := patternfactory.Parse(log)
	if parsed == nil {
		t.Error("Expected not nil but got nil")
	}

	grok, ok := parsed.Labels[(string(model.DynamicLabelUsedGrok))]

	if !ok {
		t.Error("Can't find pattern")
	}

	if grok != model.MetaLog_LogFmt.String() {
		t.Errorf("Expected pattern [%s] but got [%s]", model.MetaLog_LogFmt.String(), grok)
	}

	if len(parsed.ProcessError.Reason) == 0 {
		t.Error("Expect a process error bu got nothing")
	}

	if len(parsed.Message) == 0 {
		t.Error("Expected a message nothing")
	}

}

func TestNomadLog(t *testing.T) {
	log := TestMetaLogFromJournalDFromConst([]byte(testJournaldNomadLog), t)
	parsed := patternfactory.Parse(log)
	if parsed == nil {
		t.Error("Expected not nil but got nil")
	}

	grok, ok := parsed.Labels[(string(model.DynamicLabelUsedGrok))]

	if !ok {
		t.Error("Can't find pattern")
	}

	if grok != model.MetaLog_TsLevelMsg.String() {
		t.Errorf("Expected pattern [%s] but got [%s]", model.MetaLog_TsLevelMsg.String(), grok)
	}

	if len(parsed.ProcessError.Reason) > 0 {
		t.Errorf("Should not contain a process error. But contains %s", parsed.ProcessError.Reason)
	}

	if len(parsed.Message) == 0 {
		t.Error("Expected a message nothing")
	}

	if parsed.Log.Level != model.LogLevel_debug {
		t.Errorf("Expected a debug level but got %+v", parsed.Log.Level)
	}
	if parsed.Timestamp.AsTime().IsZero() {
		t.Errorf("Expected a valid ts but got %+v", parsed.Timestamp.AsTime())
	}

	//2023-03-20T15:06:45.057Z
	time := parsed.Timestamp.AsTime()
	year := 2023
	month := "March"
	day := 20
	hour := 15
	minute := 06
	second := 45
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
