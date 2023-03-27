package patterns

import (
	"fmt"
	"github.com/suikast42/logunifier/pkg/model"
)

type GrokPatternTsLevelMsg struct {
	GrokPatternDefault
	// Builder fields
	_parseErrors     []string
	_extractedFields map[string]string
}

func (g *GrokPatternTsLevelMsg) from(log *model.MetaLog) GrokPatternExtractor {
	compilerFor := Instance().CompilerFor(g.GrokPatternDefault.Name)
	g._this = g
	if compilerFor == nil {
		g._parseErrors = append(g._parseErrors, fmt.Sprintf("Can't a pattern for key %s", g.GrokPatternDefault.Name))
		return g._this
	}
	g._extractedFields = compilerFor.ParseString(log.Message)
	g._metaLog = log
	return g._this
}

func (g *GrokPatternTsLevelMsg) timeStamp() GrokPatternExtractor {
	panic("Not implemeneted")
	return g._this
	//tsstring, ok := g._extractedFields["timestamp"]
	//var parsedTs time.Time
	//if !ok {
	//	return g._this
	//}
	//defer func() {
	//	delete(g._extractedFields, "timestamp")
	//}()

	//cachedLayout, found := cachedLayoutForLog(g._metaLog)
	//if !found {
	//	for _, layout := range g.TimeStampFormats {
	//		_parsed, err := time.Parse(layout, tsstring)
	//		if err != nil {
	//			continue
	//		}
	//		parsedTs = _parsed
	//		cacheLayoutForLog(g._metaLog, layout)
	//	}
	//} else {
	//	_parsed, err := time.Parse(cachedLayout, tsstring)
	//	if err != nil {
	//		// Delete the ts from chance and retry all again
	//		deleteCachedLayoutForLog(g._metaLog)
	//		return g.timeStamp()
	//	}
	//	parsedTs = _parsed
	//}
	//
	//if parsedTs.IsZero() {
	//	g._parseErrors = append(g._parseErrors, fmt.Sprintf("Can't find timestamp for %s", tsstring))
	//	return g._this
	//}
	//g._timeStamp = timestamppb.New(parsedTs)

	return g._this
}

func (g *GrokPatternTsLevelMsg) message() GrokPatternExtractor {

	message, ok := g._extractedFields["message"]
	if !ok {
		return g._this
	}
	defer func() {
		delete(g._extractedFields, "message")
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

	level, levelFound := g._extractedFields["level"]
	if levelFound {
		defer func() {
			delete(g._extractedFields, "level")
		}()
		g._logInfo.Level = model.StringToLogLevel(level)
		g._logInfo.LevelEmoji = model.LogLevelToEmoji(g._logInfo.Level)
	}
	return g._this

}
