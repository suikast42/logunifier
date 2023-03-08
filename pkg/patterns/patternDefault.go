package patterns

import (
	"github.com/suikast42/logunifier/pkg/model"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GrokPatternDefault struct {
	GrokPattern
	// Builder fields
	_timeStamp        timestamppb.Timestamp
	_metaLog          *model.MetaLog
	_tags             []string
	_labels           map[string]string
	_message          string
	_logInfo          *model.Log
	_containerInfo    *model.Container
	_agentInfo        *model.Agent
	_hostInfo         *model.Host
	_organisationInfo *model.Organization
	_serviceInfo      *model.Service
	_errorInfo        *model.Error
	_traceInfo        *model.Tracing
}

func (g *GrokPatternDefault) from(log *model.MetaLog) GrokPatternExtractor {
	g._metaLog = log
	return g
}

func (g *GrokPatternDefault) timeStamp() GrokPatternExtractor {
	// Copy the Fallback TS
	// We do not expect a TS information in the message
	g._timeStamp = timestamppb.Timestamp{
		Seconds: g._metaLog.FallbackTimestamp.Seconds,
		Nanos:   g._metaLog.FallbackTimestamp.Nanos,
	}
	return g
}

func (g *GrokPatternDefault) message() GrokPatternExtractor {
	g._message = g._metaLog.Message
	return g
}

func (g *GrokPatternDefault) tags() GrokPatternExtractor {
	g._tags = g._metaLog.EcsTags()
	return g
}

func (g *GrokPatternDefault) labels() GrokPatternExtractor {
	g._labels = g._metaLog.EcsLabels()
	return g
}

func (g *GrokPatternDefault) containerInfo() GrokPatternExtractor {
	g._containerInfo = g._metaLog.EcsContainerInfo()
	return g
}

func (g *GrokPatternDefault) containercontainerInfo() GrokPatternExtractor {
	g._containerInfo = g._metaLog.EcsContainerInfo()
	return g
}

func (g *GrokPatternDefault) agentInfo() GrokPatternExtractor {
	g._agentInfo = g._metaLog.EcsAgentInfo()
	return g
}

func (g *GrokPatternDefault) hostInfo() GrokPatternExtractor {
	g._hostInfo = g._metaLog.EcsHostInfo()
	return g
}

func (g *GrokPatternDefault) organisationInfo() GrokPatternExtractor {
	g._organisationInfo = g._metaLog.EcsOrganizationInfo()
	return g
}

func (g *GrokPatternDefault) serviceInfo() GrokPatternExtractor {
	g._serviceInfo = g._metaLog.EcsServiceInfo()
	return g
}

func (g *GrokPatternDefault) errorInfo() GrokPatternExtractor {
	// We do not expect a special error info in the default log pattern
	return g
}

func (g *GrokPatternDefault) logInfo() GrokPatternExtractor {
	// We do not expect a special log info in the default log pattern
	g._logInfo = &model.Log{
		File:       nil,
		Level:      g._metaLog.FallbackLoglevel,
		Logger:     "",
		ThreadName: "",
		Origin:     nil,
		Original:   "",
		Syslog:     nil,
		LevelEmoji: model.LogLevelToEmoji(g._metaLog.FallbackLoglevel),
	}
	return g
}

func (g *GrokPatternDefault) tracingInfo() GrokPatternExtractor {
	// We do not expect a special trace info in the default log pattern
	return g
}

func (g *GrokPatternDefault) extract() *model.EcsLogEntry {
	ecs := model.NewEcsLogEntry()
	ecs.Timestamp = &g._timeStamp
	ecs.Log = g._logInfo
	ecs.Message = g._message
	ecs.Labels = g._labels
	ecs.Tags = g._tags
	ecs.Container = g._containerInfo
	ecs.Agent = g._agentInfo
	ecs.Host = g._hostInfo
	ecs.Organization = g._organisationInfo
	ecs.Service = g._serviceInfo
	ecs.Error = g._errorInfo
	ecs.Trace = g._traceInfo
	ecs.ProcessError = g._metaLog.ProcessError
	return ecs
}