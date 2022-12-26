package journald

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"github.com/nats-io/nats.go"
	"github.com/suikast42/logunifier/internal/config"
	"github.com/suikast42/logunifier/internal/streams/egress"
	"github.com/suikast42/logunifier/internal/streams/ingress"
	"github.com/suikast42/logunifier/pkg/model"
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

func NewSubscription(name string, durableSubscriptionName string, subscription string, pushChannel chan<- *egress.MsgContext) (*ingress.IngresSubscription, error) {
	logger := config.Logger()
	cfg, err := config.Instance()
	if err != nil {
		logger.Error().Err(err).Msgf("Can't obtain config in NewSubscription for  %s", name)
		return nil, err
	}

	//stream cfg
	streamCfg, err := cfg.IngressSubscription(name, "Ingress channel for journald logs comes over vector", []string{cfg.IngresNatsTest()})

	if err != nil {
		logger.Error().Err(err).Msgf("Can't create stream config %s", name)
		return nil, err
	}
	return ingress.NewIngresSubscription(durableSubscriptionName, name, subscription, &logger, pushChannel, &JournaldDToEcsConverter{}, streamCfg), nil
}

func (r *JournaldDToEcsConverter) Convert(msg *nats.Msg) *model.EcsLogEntry {
	journald := IngressSubjectJournald{}
	err := json.Unmarshal(msg.Data, &journald)
	if err != err {
		return model.ToUnmarshalError(msg, err)

	}
	return &model.EcsLogEntry{
		Id:        journald.UID,
		Message:   journald.Message,
		Timestamp: timestamppb.New(journald.Timestamp),
		Host: &model.Host{
			Hostname: journald.Host,
			Id:       journald.MACHINEID,
		},
	}

}
func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}
