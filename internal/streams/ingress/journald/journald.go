package journald

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"github.com/nats-io/nats.go"
	"github.com/suikast42/logunifier/internal/config"
	"github.com/suikast42/logunifier/internal/streams/ingress"
	"github.com/suikast42/logunifier/pkg/model"
	"github.com/suikast42/logunifier/pkg/patterns"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strings"
	"time"
)

type JournaldDToEcsConverter struct {
}
type IngressSubjectJournald struct {
	COM_HASHICORP_NOMAD_ALLOC_ID        string    `json:"COM_HASHICORP_NOMAD_ALLOC_ID"`
	COM_HASHICORP_NOMAD_JOB_ID          string    `json:"COM_HASHICORP_NOMAD_JOB_ID"`
	COM_HASHICORP_NOMAD_JOB_NAME        string    `json:"COM_HASHICORP_NOMAD_JOB_NAME"`
	COM_HASHICORP_NOMAD_NAMESPACE       string    `json:"COM_HASHICORP_NOMAD_NAMESPACE"`
	COM_HASHICORP_NOMAD_NODE_ID         string    `json:"COM_HASHICORP_NOMAD_NODE_ID"`
	COM_HASHICORP_NOMAD_NODE_NAME       string    `json:"COM_HASHICORP_NOMAD_NODE_NAME"`
	COM_HASHICORP_NOMAD_TASK_GROUP_NAME string    `json:"COM_HASHICORP_NOMAD_TASK_GROUP_NAME"`
	COM_HASHICORP_NOMAD_TASK_NAME       string    `json:"COM_HASHICORP_NOMAD_TASK_NAME"`
	CONTAINER_ID                        string    `json:"CONTAINER_ID"`
	CONTAINER_ID_FULL                   string    `json:"CONTAINER_ID_FULL"`
	CONTAINER_NAME                      string    `json:"CONTAINER_NAME"`
	CONTAINER_TAG                       string    `json:"CONTAINER_TAG"`
	CONTAINER_PARTIAL_MESSAGE           string    `json:"CONTAINER_PARTIAL_MESSAGE"`
	IMAGE_NAME                          string    `json:"IMAGE_NAME"`
	ORG_OPENCONTAINERS_IMAGE_REVISION   string    `json:"ORG_OPENCONTAINERS_IMAGE_REVISION"`
	ORG_OPENCONTAINERS_IMAGE_SOURCE     string    `json:"ORG_OPENCONTAINERS_IMAGE_SOURCE"`
	ORG_OPENCONTAINERS_IMAGE_TITLE      string    `json:"ORG_OPENCONTAINERS_IMAGE_TITLE"`
	PRIORITY                            string    `json:"PRIORITY"`
	SYSLOG_IDENTIFIER                   string    `json:"SYSLOG_IDENTIFIER"`
	BOOTID                              string    `json:"_BOOT_ID"`
	CAPEFFECTIVE                        string    `json:"_CAP_EFFECTIVE"`
	CMDLINE                             string    `json:"_CMDLINE"`
	COMM                                string    `json:"_COMM"`
	EXE                                 string    `json:"_EXE"`
	GID                                 string    `json:"_GID"`
	MACHINEID                           string    `json:"_MACHINE_ID"`
	PID                                 string    `json:"_PID"`
	SELINUXCONTEXT                      string    `json:"_SELINUX_CONTEXT"`
	STREAMID                            string    `json:"_STREAM_ID"`
	SOURCEREALTIMETIMESTAMP             string    `json:"_SOURCE_REALTIME_TIMESTAMP"`
	SYSTEMDCGROUP                       string    `json:"_SYSTEMD_CGROUP"`
	SYSTEMDINVOCATIONID                 string    `json:"_SYSTEMD_INVOCATION_ID"`
	SYSTEMDSLICE                        string    `json:"_SYSTEMD_SLICE"`
	SYSTEMDUNIT                         string    `json:"_SYSTEMD_UNIT"`
	TRANSPORT                           string    `json:"_TRANSPORT"`
	UID                                 string    `json:"_UID"`
	MONOTONICTIMESTAMP                  string    `json:"__MONOTONIC_TIMESTAMP"`
	REALTIMETIMESTAMP                   string    `json:"__REALTIME_TIMESTAMP"`
	Host                                string    `json:"host"`
	Message                             string    `json:"message"`
	SourceType                          string    `json:"source_type"`
	Timestamp                           time.Time `json:"timestamp"`
}

var unitToPattern = map[string]patterns.PatternKey{
	"init.scope":   patterns.NopPattern,
	"cron.service": patterns.NopPattern,
	"keycloak":     patterns.KeyCloakPattern,
	"nexus":        patterns.CommonUtcPatternWithCommaTsAndTz,
	//"logunifier":   patterns.CommonPatternNano,
}

func (r *JournaldDToEcsConverter) Convert(msg *nats.Msg) *model.EcsLogEntry {
	journald := IngressSubjectJournald{}
	err := json.Unmarshal(msg.Data, &journald)
	if err != err {
		return model.ToUnmarshalError(msg, err)
	}
	patternKey := journald.patternKey()
	pattern, patternFound := unitToPattern[patternKey]
	if !patternFound {
		if journald.isContainerLog() {
			// TODO: This ugly patterns must solved dynamic
			// https://github.com/suikast42/logunifier/issues/5
			if strings.HasPrefix(patternKey, "connect-proxy-") {
				// Consul connect have dynamic container name with connect-proxy- prefix
				unitToPattern[patternKey] = patterns.ConsulConnectPattern
				pattern, patternFound = unitToPattern[patternKey]
			} else if strings.HasSuffix(patternKey, "postgres") {
				// the keycloak security postgres task x_postgres
				unitToPattern[patternKey] = patterns.ConsulConnectPattern
				pattern, patternFound = unitToPattern[patternKey]
			} else {
				unitToPattern[patternKey] = patterns.CommonPattern
				pattern, patternFound = unitToPattern[patternKey]
			}
		} else {
			pattern = patterns.CommonPattern
		}
	}
	var parsed patterns.ParseResult
	// A registered pattern found for message
	def := patterns.ParseResult{
		LogLevel:  model.LogLevel_unknown,
		TimeStamp: journald.Timestamp,
	}
	parsed, err = patterns.Instance().ParseWitDefaults("IngressSubjectJournald", patternKey, def, pattern, journald.Message)
	if err != nil {
		return model.ToUnmarshalError(msg, err)
	}

	converted := &model.EcsLogEntry{
		Id:      model.UUID(),
		Message: journald.Message,
		Labels: map[string]string{
			ingress.IndexedLabelIngress:     "vector-journald",
			ingress.IndexedLabelUsedPattern: parsed.UsedPattern,
			ingress.IndexedLabelJob:         journald.jobName(),
			ingress.IndexedLabelJobType:     journald.jobType(),
		},
		Timestamp: timestamppb.New(parsed.TimeStamp),
		Tags:      []string{journald.SourceType},
		Log: &model.Log{
			Level: parsed.LogLevel,
		},
		Host: &model.Host{
			Hostname: journald.Host,
			Id:       journald.MACHINEID,
		},
		Service: &model.Service{
			EphemeralId: journald.BOOTID,
			Name:        journald.jobName(),
			Node: &model.Service_Node{
				Name: journald.Host,
			},
			Type: journald.jobType(),
		},
	}
	if journald.isContainerLog() {
		containerLabels := map[string]string{
			ingress.IndexedContainerLabelStackName: journald.COM_HASHICORP_NOMAD_JOB_NAME,
			ingress.IndexedContainerLabelTaskGroup: journald.COM_HASHICORP_NOMAD_TASK_GROUP_NAME,
			ingress.IndexedContainerLabelTask:      patternKey,
			ingress.IndexedContainerLabelNamespace: journald.COM_HASHICORP_NOMAD_NAMESPACE,
		}
		converted.Container = &model.Container{
			Id: journald.CONTAINER_ID,
			Image: &model.Container_Image{
				Name: journald.IMAGE_NAME,
			},
			Labels: containerLabels,
			Name:   journald.CONTAINER_NAME,
		}
		if len(journald.CONTAINER_PARTIAL_MESSAGE) > 0 {
			converted.Message = converted.Message + "\n" + journald.CONTAINER_PARTIAL_MESSAGE
		}
	}
	return converted

}

func (r *IngressSubjectJournald) patternKey() string {
	if len(r.COM_HASHICORP_NOMAD_TASK_NAME) > 0 {
		return r.COM_HASHICORP_NOMAD_TASK_NAME
	}
	if len(r.CONTAINER_NAME) > 0 && len(r.COM_HASHICORP_NOMAD_TASK_NAME) == 0 {
		logger := config.Logger()
		logger.Warn().Msgf("CONTAINER_NAME %s specified but cannot find  COM_HASHICORP_NOMAD_TASK_NAME. Use SYSTEMDUNIT as pattern key. ", r.CONTAINER_NAME)
	}
	return r.SYSTEMDUNIT
}

func (r *IngressSubjectJournald) jobName() string {
	if r.isContainerLog() {
		return r.patternKey()
	}
	return r.SYSTEMDUNIT
}

func (r *IngressSubjectJournald) jobType() string {
	if len(r.COM_HASHICORP_NOMAD_ALLOC_ID) > 0 {
		return "nomad_job"
	}
	if len(r.CONTAINER_NAME) > 0 {
		return "container"
	}
	return "os_service"
}
func (r *IngressSubjectJournald) isContainerLog() bool {
	return len(r.CONTAINER_NAME) > 0
}
func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}
