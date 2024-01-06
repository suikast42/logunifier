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

func (ecs *EcsLogEntry) IsJobNameSet() bool {
	return len(ecs.Labels[string(StaticLabelJob)]) > 0
}

func (ecs *EcsLogEntry) JobName() string {
	return ecs.Labels[string(StaticLabelJob)]
}

func (ecs *EcsLogEntry) SetJobName(jobName string) {
	ecs.Labels[string(StaticLabelJob)] = jobName
}

func (ecs *EcsLogEntry) IsJobTypeSet() bool {
	return len(ecs.Labels[string(StaticLabelJobType)]) > 0
}

func (ecs *EcsLogEntry) JobType() string {
	return ecs.Labels[string(StaticLabelJobType)]
}

func (ecs *EcsLogEntry) SetJobType(jobType string) {
	ecs.Labels[string(StaticLabelJobType)] = jobType
}

func (ecs *EcsLogEntry) IsPatternSet() bool {
	return len(ecs.Labels[string(DynamicLabelUsedGrok)]) > 0
}

func (ecs *EcsLogEntry) SetPattern(pattern string) {
	ecs.Labels[string(DynamicLabelUsedGrok)] = pattern
}

func (ecs *EcsLogEntry) SetLogLevel(level LogLevel) {
	if ecs.Log == nil {
		ecs.Log = &Log{}
	}
	ecs.Log.Level = level
	ecs.Log.LevelEmoji = LogLevelToEmoji(level)
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
	if len(ecs.ProcessError.Reason) == 0 {
		ecs.ProcessError.Reason = _error
	} else {
		ecs.ProcessError.Reason = ecs.ProcessError.Reason + ",\n" + _error
	}
}

func (log *MetaLog) AppendParseError(_error string) {
	if len(_error) == 0 {
		return
	}
	if len(log.ProcessError.Reason) == 0 {
		log.ProcessError.Reason = _error
	} else {
		log.ProcessError.Reason = log.ProcessError.Reason + ",\n" + _error
	}
}
func (log *MetaLog) HasProcessErrors() bool {
	return log.ProcessError != nil && log.ProcessError.Reason != ""
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
	// The protojson package does not tollerate the UnmarshalJSON logic of json deserilisation
	// so that we can't do a log level mapping from uppercased log levels fpr example
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
