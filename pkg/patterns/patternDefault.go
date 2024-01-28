package patterns

import (
	"github.com/suikast42/logunifier/pkg/model"
	"strings"
)

type GrokPatternDefault struct {
	GrokPattern
	// region Builder fields
	// this reference set by initial creation with from the
	_this        GrokPatternExtractor
	_metaLog     *model.MetaLog
	_parseErrors []string
	//endregion
}

func (g *GrokPatternDefault) from(log *model.MetaLog) GrokPatternExtractor {
	g._metaLog = log
	g._this = g
	return g._this
}

func (g *GrokPatternDefault) timeStamp() GrokPatternExtractor {
	// Use the provided timestamp from ingress
	return g._this
}

func (g *GrokPatternDefault) message() GrokPatternExtractor {
	// We have no message pattern. Copy the raw
	g._metaLog.EcsLogEntry.Message = g._metaLog.RawMessage
	return g._this
}

func (g *GrokPatternDefault) tags() GrokPatternExtractor {
	return g._this
}

func (g *GrokPatternDefault) labels() GrokPatternExtractor {
	return g._this
}

func (g *GrokPatternDefault) containerInfo() GrokPatternExtractor {
	return g._this
}

func (g *GrokPatternDefault) agentInfo() GrokPatternExtractor {
	return g._this
}

func (g *GrokPatternDefault) hostInfo() GrokPatternExtractor {
	return g._this
}

func (g *GrokPatternDefault) organisationInfo() GrokPatternExtractor {
	return g._this
}

func (g *GrokPatternDefault) serviceInfo() GrokPatternExtractor {
	return g._this
}

func (g *GrokPatternDefault) errorInfo() GrokPatternExtractor {
	// We do not expect a special error info in the default log pattern
	return g._this
}

func (g *GrokPatternDefault) eventInfo() GrokPatternExtractor {
	// We do not expect a special error info in the default log pattern
	return g._this
}

func (g *GrokPatternDefault) logInfo() GrokPatternExtractor {
	//We don't know the log level
	g._metaLog.EcsLogEntry.SetLogLevel(model.LogLevel_unknown)
	return g._this
}

func (g *GrokPatternDefault) tracingInfo() GrokPatternExtractor {
	// We do not expect a special trace info in the default log pattern
	return g._this
}

func (g *GrokPatternDefault) userInfo() GrokPatternExtractor {
	// We do not expect a special user info in the default log pattern
	return g._this
}

func (g *GrokPatternDefault) extract() *model.EcsLogEntry {
	ecs := g._metaLog.EcsLogEntry
	if len(g._parseErrors) > 0 {
		ecs.AppendParseError(strings.Join(g._parseErrors, "\n"))
	}
	return ecs
}
