package patterns

import (
	"github.com/suikast42/logunifier/pkg/model"
)

type GrokPatternKey string

// GrokPattern A Grok pattern definition for parsing a log line
type GrokPattern struct {

	// Pattern the pattern used for parse the log message
	Name model.MetaLog_PatternKey
}

type GrokPatternExtractor interface {
	// From Extract the log information form this model.MetaLog
	from(log *model.MetaLog) GrokPatternExtractor

	// TimeStamp The timestamp either from model.MetaLog or from the fallback timestamp of model.MetaLog
	timeStamp() GrokPatternExtractor

	// Message extract the message from model.MetaLog
	message() GrokPatternExtractor

	//Tags Extract the tags from model.MetaLog
	tags() GrokPatternExtractor

	//Labels Extract the labels from model.MetaLog
	labels() GrokPatternExtractor

	// ContainerInfo If the log source is a container then fill its meta information from here
	containerInfo() GrokPatternExtractor

	// AgentInfo The agent information of the client
	// For example filebeat or a java client that ships the logs
	agentInfo() GrokPatternExtractor

	// HostInfo Host information of the logging agent.
	hostInfo() GrokPatternExtractor

	// OrganisationInfo Information about the project
	organisationInfo() GrokPatternExtractor

	// ServiceInfo Service information
	serviceInfo() GrokPatternExtractor

	// ErrorInfo Log error information
	errorInfo() GrokPatternExtractor

	// LogInfo Meta information about the log
	logInfo() GrokPatternExtractor

	//TracingInfo Apm trace
	tracingInfo() GrokPatternExtractor

	// Information about the user in a log entry
	userInfo() GrokPatternExtractor

	// Information about the kind of a log event
	// eg. alert, enrichment, event, metric, state, pipeline_error, signal
	eventInfo() GrokPatternExtractor

	// Create finally the model.EcsLogEntry
	extract() *model.EcsLogEntry
}

func ExtractFrom(extractor GrokPatternExtractor, from *model.MetaLog) *model.EcsLogEntry {

	return extractor.from(from).
		timeStamp().
		message().
		tags().
		labels().
		containerInfo().
		agentInfo().
		hostInfo().
		organisationInfo().
		serviceInfo().
		errorInfo().
		logInfo().
		tracingInfo().
		extract()
}
