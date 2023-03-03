package patterns

import (
	"github.com/suikast42/logunifier/pkg/model"
	"github.com/trivago/grok"
)

type GrokPatternKey string

// GrokPattern A Grok pattern definition for parsing a log line
type GrokPattern struct {

	// Name unique name of the pattern
	GrokPatternKey string

	// Pattern the pattern used for parse the log message
	Name GrokPatternKey

	// CompiledPattern the compiled grok pattern from Pattern
	CompiledPattern *grok.CompiledGrok

	// TimeStampFormat the timestamp format
	// for example time.RFC3339
	TimeStampFormat string
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

	// Create finally the model.EcsLogEntry
	extract() *model.EcsLogEntry
}

func ExractFrom(extractor GrokPatternExtractor, from *model.MetaLog) *model.EcsLogEntry {
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
