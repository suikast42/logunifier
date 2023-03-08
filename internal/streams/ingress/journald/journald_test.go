package journald

import (
	"github.com/rs/zerolog"
	"github.com/suikast42/logunifier/internal/config"
	"github.com/suikast42/logunifier/pkg/patterns"
	"github.com/suikast42/logunifier/pkg/utils"
	"os"
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

}
