package model

import (
	"bytes"
	"encoding/json"
	"google.golang.org/protobuf/types/known/timestamppb"
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
