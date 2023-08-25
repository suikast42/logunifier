package patterns

import (
	"github.com/suikast42/logunifier/pkg/model"
	"strings"
)

type GrokPatternEcs struct {
	GrokPatternDefault
	// Builder fields

	_parsedEcs *model.EcsLogEntry
}

func (g *GrokPatternEcs) from(log *model.MetaLog) GrokPatternExtractor {
	g._metaLog = log
	g._this = g
	g._parsedEcs = model.NewEcsLogEntry()
	err := g._parsedEcs.FromJson([]byte(log.Message))
	if err != nil {
		g._parseErrors = append(g._parseErrors, err.Error())
		//return g._this
	}

	return g._this
}

func (g *GrokPatternEcs) timeStamp() GrokPatternExtractor {
	// ecs contains ts itself
	if g.hasParseError() {
		g._parsedEcs.SetTimeStamp(g._metaLog.FallbackTimestamp)
	}
	return g._this
}

func (g *GrokPatternEcs) message() GrokPatternExtractor {
	if g.hasParseError() {
		g._parsedEcs.Message = g._metaLog.Message
	}
	return g._this
}

func (g *GrokPatternEcs) tags() GrokPatternExtractor {
	if len(g._metaLog.Tags) > 0 {
		g._parsedEcs.Tags = append(g._parsedEcs.Tags, g._metaLog.Tags...)
	}
	return g._this
}

func (g *GrokPatternEcs) labels() GrokPatternExtractor {
	// Merge labels from metalog
	for k, v := range g._metaLog.EcsLabels() {
		g._parsedEcs.Labels[k] = v
	}
	g._parsedEcs.Labels[string(model.DynamicLabelUsedGrok)] = g.Name.String()
	return g._this
}

func (g *GrokPatternEcs) containerInfo() GrokPatternExtractor {
	info := g._metaLog.EcsContainerInfo()
	if info != nil {
		g._parsedEcs.Container = info
	}
	return g._this
}

func (g *GrokPatternEcs) agentInfo() GrokPatternExtractor {
	info := g._metaLog.EcsAgentInfo()
	if info != nil {
		g._parsedEcs.Agent = info
	}
	return g._this
}

func (g *GrokPatternEcs) hostInfo() GrokPatternExtractor {
	info := g._metaLog.EcsHostInfo()
	if info != nil {
		g._parsedEcs.Host = info
	}
	return g._this
}

func (g *GrokPatternEcs) organisationInfo() GrokPatternExtractor {
	info := g._metaLog.EcsOrganizationInfo()
	if info != nil {
		g._parsedEcs.Organization = info
	}
	return g._this
}

func (g *GrokPatternEcs) serviceInfo() GrokPatternExtractor {
	info := g._metaLog.EcsServiceInfo()
	if info != nil {
		g._parsedEcs.Service = info
	}
	return g._this
}

func (g *GrokPatternEcs) errorInfo() GrokPatternExtractor {
	// We do not expect a special error info in the default log pattern
	return g._this
}

func (g *GrokPatternEcs) eventInfo() GrokPatternExtractor {
	// We do not expect a special event info in the default log pattern
	return g._this
}

func (g *GrokPatternEcs) logInfo() GrokPatternExtractor {
	//We do not expect a special log info from ingress type
	return g._this
}

func (g *GrokPatternEcs) tracingInfo() GrokPatternExtractor {
	// We do not expect a special trace info in the default log pattern
	return g._this
}

func (g *GrokPatternEcs) userInfo() GrokPatternExtractor {
	// We do not expect a special user info in the default log pattern
	return g._this
}

func (g *GrokPatternEcs) extract() *model.EcsLogEntry {
	if len(g._parseErrors) > 0 {
		g._parsedEcs.AppendParseError(strings.Join(g._parseErrors, "\n"))
	}
	return g._parsedEcs
}
