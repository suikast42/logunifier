package journald

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"github.com/nats-io/nats.go"
	"github.com/suikast42/logunifier/pkg/model"
	"github.com/suikast42/logunifier/pkg/patterns"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

type JournaldDToEcsConverter struct {
}
type IngressSubjectJournald struct {
	PRIORITY            string    `json:"PRIORITY"`
	SYSLOGFACILITY      string    `json:"SYSLOG_FACILITY"`
	SYSLOGIDENTIFIER    string    `json:"SYSLOG_IDENTIFIER"`
	BOOTID              string    `json:"_BOOT_ID"`
	CAPEFFECTIVE        string    `json:"_CAP_EFFECTIVE"`
	CMDLINE             string    `json:"_CMDLINE"`
	COMM                string    `json:"_COMM"`
	EXE                 string    `json:"_EXE"`
	GID                 string    `json:"_GID"`
	MACHINEID           string    `json:"_MACHINE_ID"`
	PID                 string    `json:"_PID"`
	SELINUXCONTEXT      string    `json:"_SELINUX_CONTEXT"`
	STREAMID            string    `json:"_STREAM_ID"`
	SYSTEMDCGROUP       string    `json:"_SYSTEMD_CGROUP"`
	SYSTEMDINVOCATIONID string    `json:"_SYSTEMD_INVOCATION_ID"`
	SYSTEMDSLICE        string    `json:"_SYSTEMD_SLICE"`
	SYSTEMDUNIT         string    `json:"_SYSTEMD_UNIT"`
	TRANSPORT           string    `json:"_TRANSPORT"`
	UID                 string    `json:"_UID"`
	MONOTONICTIMESTAMP  string    `json:"__MONOTONIC_TIMESTAMP"`
	REALTIMETIMESTAMP   string    `json:"__REALTIME_TIMESTAMP"`
	Host                string    `json:"host"`
	Message             string    `json:"message"`
	SourceType          string    `json:"source_type"`
	Timestamp           time.Time `json:"timestamp"`
}

var unitToPattern map[string]patterns.PatternKey

func init() {
	unitToPattern = make(map[string]patterns.PatternKey)
	unitToPattern["nomad.service"] = patterns.TS_LEVEL_MSG
	unitToPattern["consul.service"] = patterns.TS_LEVEL_MSG
}
func (r *JournaldDToEcsConverter) Convert(msg *nats.Msg) *model.EcsLogEntry {
	journald := IngressSubjectJournald{}
	err := json.Unmarshal(msg.Data, &journald)
	if err != err {
		return model.ToUnmarshalError(msg, err)
	}
	pattern, patternFound := unitToPattern[journald.SYSTEMDUNIT]
	var parsed patterns.ParseResult
	// A registered pattern found for message
	def := patterns.ParseResult{
		LogLevel:  "UNKNOWN",
		TimeStamp: journald.Timestamp,
		Msg:       journald.Message,
	}
	if patternFound {
		parsed, err = patterns.Instance().ParseWitDefaults(def, pattern, journald.Message)
		if err != nil {
			return model.ToUnmarshalError(msg, err)
		}
	} else {
		parsed = def
	}
	return &model.EcsLogEntry{
		Id:        model.UUID(),
		Message:   parsed.Msg,
		Timestamp: timestamppb.New(parsed.TimeStamp),
		Log: &model.Log{
			Level: parsed.LogLevel,
		},
		Host: &model.Host{
			Hostname: journald.Host,
			Id:       journald.MACHINEID,
		},
		Service: &model.Service{
			EphemeralId: journald.BOOTID,
			Name:        journald.SYSTEMDUNIT,
			Node: &model.Service_Node{
				Name: journald.Host,
			},
			Type: journald.SourceType,
		},
	}

}
func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}
