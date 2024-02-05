package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/xyproto/jpath"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strconv"
	"strings"
	"time"
)

func (ecs *EcsLogEntry) HasProcessError() bool {
	return ecs.ProcessError != nil && len(ecs.ProcessError.Reason) > 0
}

func (ecs *EcsLogEntry) HasValidationError() bool {
	return ecs.ValidationError != nil && len(ecs.ValidationError.Errors) > 0
}

func (ecs *EcsLogEntry) HasExceptionStackStrace() bool {
	return ecs.Error != nil && len(ecs.Error.StackTrace) > 0
}

func (ecs *EcsLogEntry) HasTags() bool {
	return ecs.Tags != nil && len(ecs.Tags) > 0
}

func (ecs *EcsLogEntry) IsIngressSet() bool {
	return ecs.Log != nil && len(ecs.Log.Ingress) > 0
}
func (ecs *EcsLogEntry) IsServiceNameSet() bool {
	return ecs.Service != nil && len(ecs.Service.Name) > 0
}

func (ecs *EcsLogEntry) SetSetServiceName(serviceName string) {
	if ecs.Service != nil {
		ecs.Service = &Service{}
	}
	ecs.Service.Name = serviceName
}

func (ecs *EcsLogEntry) SetIngress(ingress string) {
	if ecs.Log != nil {
		ecs.Log = &Log{}
	}
	ecs.Log.Ingress = ingress
}

func (ecs *EcsLogEntry) IsServiceTypeSet() bool {
	return ecs.Service != nil && len(ecs.Service.Type) > 0
}

func (ecs *EcsLogEntry) IsLogLevelSet() bool {
	return ecs.Log != nil && ecs.Log.Level != LogLevel_not_set
}
func (ecs *EcsLogEntry) SetServiceType(serviceType string) {
	if ecs.Service == nil {
		ecs.Service = &Service{}
	}
	ecs.Service.Type = serviceType
}

func (ecs *EcsLogEntry) IsPatternSet() bool {
	return ecs.Log != nil && len(ecs.Log.PatternKey) > 0
}

func (ecs *EcsLogEntry) IsOrgNameSet() bool {
	return ecs.Organization != nil && len(ecs.Organization.Name) > 0
}

func (ecs *EcsLogEntry) IsEnvironmentSet() bool {
	return ecs.Environment != nil && len(ecs.Environment.Name) > 0
}

func (ecs *EcsLogEntry) IsStackSet() bool {
	return ecs.Service != nil && len(ecs.Service.Stack) > 0
}

func (ecs *EcsLogEntry) IsServiceNameSpaceSet() bool {
	return ecs.Service != nil && len(ecs.Service.Namespace) > 0
}

func (ecs *EcsLogEntry) SetServiceNameSpace(namespace string) {
	if ecs.Service == nil {
		ecs.Service = &Service{}
	}
	ecs.Service.Namespace = namespace
}
func (ecs *EcsLogEntry) SetStack(stackname string) {
	if ecs.Service == nil {
		ecs.Service = &Service{}
	}
	ecs.Service.Stack = stackname
}

func (ecs *EcsLogEntry) IsHostNameSet() bool {
	return ecs.Host != nil && len(ecs.Host.Name) > 0 && len(ecs.Host.Hostname) > 0
}

func (ecs *EcsLogEntry) IsTraceIdSet() bool {
	return ecs.Trace != nil && ecs.Trace.Trace != nil && len(ecs.Trace.Trace.Id) > 0
}

func (ecs *EcsLogEntry) IsSpanIdSet() bool {
	return ecs.Trace != nil && ecs.Trace.Span != nil && len(ecs.Trace.Span.Id) > 0
}

func (ecs *EcsLogEntry) SetHostName(hostName string) {
	if ecs.Host == nil {
		ecs.Host = &Host{}
	}
	ecs.Host.Name = hostName
	ecs.Host.Hostname = hostName
}

func (ecs *EcsLogEntry) SetPattern(pattern string) {
	if ecs.Log == nil {
		ecs.Log = &Log{}
	}
	ecs.Log.PatternKey = pattern
}

func (ecs *EcsLogEntry) SetOrgName(orgName string) {
	if ecs.Organization == nil {
		ecs.Organization = &Organization{}
	}
	ecs.Organization.Name = orgName
	ecs.Organization.Id = "0"
}

func (ecs *EcsLogEntry) SetEnvironment(environment string) {
	if ecs.Environment == nil {
		ecs.Environment = &Environment{}
	}
	ecs.Environment.Name = environment
}

func (ecs *EcsLogEntry) SetLogLevel(level LogLevel) {
	if ecs.Log == nil {
		ecs.Log = &Log{}
	}
	ecs.Log.Level = level
	ecs.Log.LevelEmoji = LogLevelToEmoji(level)

}

func (ecs *EcsLogEntry) SetMarkerEmojis() {
	if ecs.HasTags() {
		ecs.Log.LevelEmoji = ecs.Log.LevelEmoji + " " + EmojiMarker()
	}

	if ecs.HasExceptionStackStrace() {
		ecs.Log.LevelEmoji = ecs.Log.LevelEmoji + " " + EmojiStackStrace()
	}

}

func (ecs *EcsLogEntry) SetMarkerApm() {
	if ecs.IsTraceIdSet() {
		ecs.Log.LevelEmoji = ecs.Log.LevelEmoji + " " + ApmMarker()
	}
}

func (ecs *EcsLogEntry) SetTimeStamp(timestamp *timestamppb.Timestamp) {
	ecs.Timestamp = timestamp
}

func (ecs *EcsLogEntry) GetTimeStamp() time.Time {
	if ecs.Timestamp != nil {
		return ecs.Timestamp.AsTime()
	}

	return time.Time{}
}

func (ecs *EcsLogEntry) AppendParseError(_error string) {
	if len(_error) == 0 {
		return
	}
	if ecs.ProcessError == nil {
		ecs.ProcessError = &ProcessError{
			Reason:  "",
			RawData: "",
			Subject: "Unknown",
		}
	}
	if len(ecs.ProcessError.Reason) == 0 {
		ecs.ProcessError.Reason = _error
	} else {
		ecs.ProcessError.Reason = ecs.ProcessError.Reason + ",\n" + _error
	}
}

func (ecs *EcsLogEntry) AppendValidationError(_error string) {
	if len(_error) == 0 {
		return
	}
	if ecs.ValidationError == nil {
		ecs.ValidationError = &ValidationError{}
	}
	if len(ecs.ValidationError.Errors) == 0 {
		ecs.ValidationError.Errors = _error
	} else {
		ecs.ValidationError.Errors = ecs.ValidationError.Errors + ",\n" + _error
	}

}

func (log *MetaLog) AppendParseError(_error string) {
	if len(_error) == 0 {
		return
	}
	if len(log.EcsLogEntry.ProcessError.Reason) == 0 {
		log.EcsLogEntry.ProcessError.Reason = _error
	} else {
		log.EcsLogEntry.ProcessError.Reason = log.EcsLogEntry.ProcessError.Reason + ",\n" + _error
	}
}
func (log *MetaLog) HasProcessErrors() bool {
	return log.EcsLogEntry.ProcessError != nil && log.EcsLogEntry.ProcessError.Reason != ""
}

// MarshalJSON Json serializes for log level enum
func (s LogLevel) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(LogLevelToString(s))
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

// UnmarshalJSON Json deserializes for log level enum
func (s *LogLevel) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	*s = StringToLogLevel(j)
	return nil
}

func (s LogLevel) Marshal() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(LogLevelToString(s))
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

// UnmarshalJSON Json deserializes for log level enum
func (s *LogLevel) Unmarshal(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	*s = StringToLogLevel(j)
	return nil
}

func (ecs *EcsLogEntry) ToJson() ([]byte, error) {
	// Use proto json for serialization
	// the regular json package does not tolerate the json annotations of protobuf
	protoJson, err := protojson.Marshal(ecs)
	if err != nil {
		return []byte{}, err
	}

	return protoJson, nil
}

func (ecs *EcsLogEntry) FromJson(jsonString []byte) error {
	// TODO: https://github.com/suikast42/logunifier/issues/26
	// The protojson package does not tolerate the UnmarshalJSON logic of json deserialization
	// so that we can't do a log level mapping from uppercase log levels fpr example
	// The native json package works well with the UnmarshalJSON logic but can't handle custom
	// timestamp formats.
	// We can't UnmarshalJSON for proto timestamp because that package is not in our control
	// So we let do the unmarshal logic his work do and convert afterward only the timestamp field

	node, err := jpath.New(jsonString)
	if err != nil {
		return err
	}

	err = json.Unmarshal(node.MustJSON(), ecs)
	if err != nil {
		// Ugly json fix. Some serializations serilise this filed as number or string
		// Our definition is string.
		if strings.Contains(err.Error(), "log.origin.file.line") {
			lineNumber := node.GetNode(".log.origin.file.line").Int()
			node.GetNode(".log.origin.file").Set("line", strconv.Itoa(lineNumber))
			errInt := json.Unmarshal(node.MustJSON(), ecs)
			if errInt != nil {
				//return the original error
				return err
			}
		} else {
			return err
		}
	}

	//m := make(map[string]any)
	//err = json.Unmarshal(jsonString, &m)
	//if err != nil {
	//	return err
	//}
	//
	//if value, ok := m["@timestamp"]; ok {
	//	parsedTs, err := time.Parse(time.RFC3339Nano, fmt.Sprintf("%v", value))
	//	if err != nil {
	//		return err
	//	}
	//	ecs.Timestamp = timestamppb.New(parsedTs)
	//}
	ts := node.GetNode("@timestamp").String()
	parsedTs, err := time.Parse(time.RFC3339Nano, fmt.Sprintf("%v", ts))
	if err != nil {
		return err
	}
	ecs.Timestamp = timestamppb.New(parsedTs)
	return nil
}

func (ecs *EcsLogEntry) FromJsonString(jsonString string) error {
	return ecs.FromJson([]byte(jsonString))
}
