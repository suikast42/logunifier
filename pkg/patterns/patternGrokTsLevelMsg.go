package patterns

import (
	"fmt"
	"github.com/suikast42/logunifier/pkg/model"
	"github.com/suikast42/logunifier/pkg/utils"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GrokPatternTsLevelMsg struct {
	GrokPatternDefault
	// Builder fields
	_extractedFields map[string]string
}

func (g *GrokPatternTsLevelMsg) from(log *model.MetaLog) GrokPatternExtractor {
	compilerFor := Instance().CompilerFor(g.GrokPatternDefault.Name)
	g._this = g
	g._metaLog = log
	if compilerFor == nil {
		g._parseErrors = append(g._parseErrors, fmt.Sprintf("Can't a pattern for key %s", g.GrokPatternDefault.Name))
		return g._this
	}
	g._extractedFields = compilerFor.ParseString(log.Message)
	return g._this
}

func (g *GrokPatternTsLevelMsg) timeStamp() GrokPatternExtractor {
	tsstring, ok := g._extractedFields[string(utils.PatternMatchTimeStamp)]
	if !ok {
		g._parseErrors = append(g._parseErrors, "Can't find timestamp")
		g._timeStamp = g._metaLog.FallbackTimestamp
		return g._this
	}
	defer func() {
		delete(g._extractedFields, string(utils.PatternMatchTimeStamp))
	}()
	parsedTs := utils.ParseTime(g._metaLog, tsstring)

	if parsedTs.IsZero() {
		g._parseErrors = append(g._parseErrors, fmt.Sprintf("Can't find timestamp for %s", tsstring))
		return g._this
	}
	g._timeStamp = timestamppb.New(parsedTs)

	return g._this
}

func (g *GrokPatternTsLevelMsg) message() GrokPatternExtractor {

	message, ok := g._extractedFields[string(utils.PatternMatchKeyMessage)]
	if !ok {
		g._parseErrors = append(g._parseErrors, "Can't find a message")
		g._message = g._metaLog.Message
		return g._this
	}
	defer func() {
		delete(g._extractedFields, string(utils.PatternMatchKeyMessage))
	}()
	g._message = message
	return g._this
}
func (g *GrokPatternTsLevelMsg) logInfo() GrokPatternExtractor {
	g._logInfo = &model.Log{
		File:       nil,
		Level:      g._metaLog.FallbackLoglevel,
		Logger:     "",
		ThreadName: "",
		Original:   "",
		Syslog:     nil,
		LevelEmoji: model.LogLevelToEmoji(g._metaLog.FallbackLoglevel),
	}

	level, levelFound := g._extractedFields[string(utils.PatternMatchKeyLevel)]
	if levelFound {
		defer func() {
			delete(g._extractedFields, string(utils.PatternMatchKeyLevel))
		}()
		g._logInfo.Level = model.StringToLogLevel(level)
		g._logInfo.LevelEmoji = model.LogLevelToEmoji(g._logInfo.Level)
	}

	return g._this

}

func (g *GrokPatternTsLevelMsg) extract() *model.EcsLogEntry {
	ecs := g.GrokPatternDefault.extract()
	// Every step removes the registered keys
	// Add the not standard keys as labels
	for k, v := range g._extractedFields {
		if utils.IsRegisteredKey(k) {
			ecs.Labels["pattern_"+k] = v
		}
	}

	return ecs
}
