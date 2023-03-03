package journald

import (
	"encoding/json"
	"github.com/nats-io/nats.go"
	"github.com/suikast42/logunifier/internal/config"
	"github.com/suikast42/logunifier/internal/streams/ingress"
	"github.com/suikast42/logunifier/pkg/model"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strconv"
	"strings"
	"time"
)

type JournaldDToEcsConverter struct {
}

// IngressSubjectJournald For journald fields see https://www.freedesktop.org/software/systemd/man/systemd.journal-fields.html
type IngressSubjectJournald struct {
	COM_HASHICORP_NOMAD_ALLOC_ID        string `json:"COM_HASHICORP_NOMAD_ALLOC_ID"`
	COM_HASHICORP_NOMAD_JOB_ID          string `json:"COM_HASHICORP_NOMAD_JOB_ID"`
	COM_HASHICORP_NOMAD_JOB_NAME        string `json:"COM_HASHICORP_NOMAD_JOB_NAME"`
	COM_HASHICORP_NOMAD_NAMESPACE       string `json:"COM_HASHICORP_NOMAD_NAMESPACE"`
	COM_HASHICORP_NOMAD_NODE_ID         string `json:"COM_HASHICORP_NOMAD_NODE_ID"`
	COM_HASHICORP_NOMAD_NODE_NAME       string `json:"COM_HASHICORP_NOMAD_NODE_NAME"`
	COM_HASHICORP_NOMAD_TASK_GROUP_NAME string `json:"COM_HASHICORP_NOMAD_TASK_GROUP_NAME"`
	COM_HASHICORP_NOMAD_TASK_NAME       string `json:"COM_HASHICORP_NOMAD_TASK_NAME"`
	COM_HASHICORP_NOMAD_NATIVE_ECS      string `json:"COM_HASHICORP_NOMAD_NATIVE_ECS"`
	CONTAINER_ID                        string `json:"CONTAINER_ID"`
	CONTAINER_ID_FULL                   string `json:"CONTAINER_ID_FULL"`
	CONTAINER_NAME                      string `json:"CONTAINER_NAME"`
	CONTAINER_TAG                       string `json:"CONTAINER_TAG"`
	//CONTAINER_PARTIAL_MESSAGE           string    `json:"CONTAINER_PARTIAL_MESSAGE"`
	IMAGE_NAME                        string    `json:"IMAGE_NAME"`
	ORG_OPENCONTAINERS_IMAGE_REVISION string    `json:"ORG_OPENCONTAINERS_IMAGE_REVISION"`
	ORG_OPENCONTAINERS_IMAGE_SOURCE   string    `json:"ORG_OPENCONTAINERS_IMAGE_SOURCE"`
	ORG_OPENCONTAINERS_IMAGE_TITLE    string    `json:"ORG_OPENCONTAINERS_IMAGE_TITLE"`
	PRIORITY                          string    `json:"PRIORITY"`
	SYSLOG_IDENTIFIER                 string    `json:"SYSLOG_IDENTIFIER"`
	BOOTID                            string    `json:"_BOOT_ID"`
	CAPEFFECTIVE                      string    `json:"_CAP_EFFECTIVE"`
	CMDLINE                           string    `json:"_CMDLINE"`
	COMM                              string    `json:"_COMM"`
	EXE                               string    `json:"_EXE"`
	GID                               string    `json:"_GID"`
	MACHINEID                         string    `json:"_MACHINE_ID"`
	PID                               string    `json:"_PID"`
	SELINUXCONTEXT                    string    `json:"_SELINUX_CONTEXT"`
	STREAMID                          string    `json:"_STREAM_ID"`
	SOURCEREALTIMETIMESTAMP           string    `json:"_SOURCE_REALTIME_TIMESTAMP"`
	SYSTEMDCGROUP                     string    `json:"_SYSTEMD_CGROUP"`
	SYSTEMDINVOCATIONID               string    `json:"_SYSTEMD_INVOCATION_ID"`
	SYSTEMDSLICE                      string    `json:"_SYSTEMD_SLICE"`
	SYSTEMDUNIT                       string    `json:"_SYSTEMD_UNIT"`
	TRANSPORT                         string    `json:"_TRANSPORT"`
	UID                               string    `json:"_UID"`
	MONOTONICTIMESTAMP                string    `json:"__MONOTONIC_TIMESTAMP"`
	REALTIMETIMESTAMP                 string    `json:"__REALTIME_TIMESTAMP"`
	Host                              string    `json:"host"`
	Message                           string    `json:"message"`
	SourceType                        string    `json:"source_type"`
	Timestamp                         time.Time `json:"timestamp"`
}

func (r *JournaldDToEcsConverter) ConvertToMetaLog(msg *nats.Msg) ingress.IngressMsgContext {
	journald := IngressSubjectJournald{}
	err := json.Unmarshal(msg.Data, &journald)
	if err != err {
		//logger := config.Logger()
		//logger.Err(err).Msgf("Can't unmarshal journald ingress.\n[%s]", string(msg.Data))
		// The parsing error is shipped to the output
		return ingress.IngressMsgContext{
			NatsMsg: msg,
			MetaLog: &model.MetaLog{
				AppVersion:        journald.appVersion(),
				PatternIdentifier: journald.patternIdentifier(),
				ParseError: &model.ParseError{
					Reason:        model.ParseError_Unmarshal,
					RawData:       string(msg.Data),
					Subject:       msg.Subject,
					MessageHeader: ingress.HeaderToMap(msg.Header),
				},
			},
		}
	}

	return ingress.IngressMsgContext{
		NatsMsg: msg,
		MetaLog: &model.MetaLog{
			AppVersion:        journald.appVersion(),
			PatternIdentifier: journald.patternIdentifier(),
			FallbackTimestamp: journald.ts(),
			FallbackLoglevel:  journald.toLogLevel(),
			Labels:            journald.extractLabels(msg),
			Tags:              journald.tags(),
			IsNativeEcs:       journald.isNativeEcs(),
			Message:           journald.Message,
			ParseError:        nil,
		},
	}
}

func (r *IngressSubjectJournald) extractLabels(msg *nats.Msg) map[string]string {
	var labels = make(map[string]string)
	labels[string(model.StaticLabelIngress)] = msg.Subject
	labels[string(model.StaticLabelJob)] = r.jobName()
	labels[string(model.StaticLabelJobType)] = string(r.jobType())
	if len(r.COM_HASHICORP_NOMAD_NAMESPACE) > 0 {
		labels[string(model.StaticLabelNameSpace)] = r.COM_HASHICORP_NOMAD_NAMESPACE
	}
	if len(r.COM_HASHICORP_NOMAD_TASK_NAME) > 0 {
		labels[string(model.StaticLabelTask)] = r.COM_HASHICORP_NOMAD_TASK_NAME
	}
	if len(r.COM_HASHICORP_NOMAD_TASK_GROUP_NAME) > 0 {
		labels[string(model.StaticLabelTaskGroup)] = r.COM_HASHICORP_NOMAD_TASK_GROUP_NAME
	}

	if len(r.CONTAINER_NAME) > 0 {
		labels[string(model.ContainerID)] = r.CONTAINER_ID
		labels[string(model.ContainerIDFull)] = r.CONTAINER_ID_FULL
		labels[string(model.ContainerName)] = r.CONTAINER_NAME
		labels[string(model.ContainerImageName)] = r.IMAGE_NAME
		labels[string(model.ContainerImageRevision)] = r.ORG_OPENCONTAINERS_IMAGE_REVISION
		labels[string(model.ContainerImageSource)] = r.ORG_OPENCONTAINERS_IMAGE_SOURCE
		labels[string(model.ContainerImageTitle)] = r.ORG_OPENCONTAINERS_IMAGE_TITLE
	}
	labels[string(model.StaticLabelHost)] = r.Host
	labels[string(model.StaticLabelHostId)] = r.MACHINEID
	return labels
}
func (r *IngressSubjectJournald) ts() *timestamppb.Timestamp {
	ts, err := strconv.ParseInt(r.REALTIMETIMESTAMP, 10, 64)
	if err != nil {
		logger := config.Logger()
		logger.Err(err).Msgf("Can't convert REALTIMETIMESTAMP [%s] to int64. Use time.Now().Unix() instead ", r.REALTIMETIMESTAMP)
		ts = time.Now().Unix()
	}
	// Convert Unix time in microseconds to time.Time object
	tm := time.Unix(0, ts*1000)
	return timestamppb.New(tm)

}
func (r *IngressSubjectJournald) toLogLevel() model.LogLevel {

	jobType := r.jobType()
	if jobType == ingress.JobTypeNomadJob || jobType == ingress.JobTypeContainer {
		return model.LogLevel_unknown
	}
	if len(r.PRIORITY) == 0 {
		return model.LogLevel_unknown
	}
	switch r.PRIORITY {
	case "0", "1", "2":
		return model.LogLevel_fatal
	case "3":
		return model.LogLevel_error
	case "4":
		return model.LogLevel_warn
	case "5", "6":
		return model.LogLevel_info
	case "7":
		return model.LogLevel_debug

	default:
		return model.LogLevel_unknown
	}
}

func (r *IngressSubjectJournald) isNativeEcs() bool {
	boolValue, err := strconv.ParseBool(r.COM_HASHICORP_NOMAD_NATIVE_ECS)
	if err != nil {
		return false
	}
	return boolValue
}

func (r *IngressSubjectJournald) jobName() string {
	if len(r.COM_HASHICORP_NOMAD_TASK_NAME) > 0 {
		return r.COM_HASHICORP_NOMAD_TASK_NAME
	}
	if len(r.CONTAINER_NAME) > 0 {
		return r.CONTAINER_NAME
	}
	return r.SYSTEMDUNIT
}

func (r *IngressSubjectJournald) jobType() ingress.JobType {
	if len(r.COM_HASHICORP_NOMAD_ALLOC_ID) > 0 {
		return ingress.JobTypeNomadJob
	}
	if len(r.CONTAINER_NAME) > 0 {
		return ingress.JobTypeContainer
	}
	return ingress.JobTypeOsService
}

func (r *IngressSubjectJournald) appVersion() string {
	return "0"
}

func (r *IngressSubjectJournald) patternIdentifier() string {
	return ""
}

func (r *IngressSubjectJournald) tags() []string {
	if len(r.CONTAINER_TAG) > 0 {
		return strings.Split(r.CONTAINER_TAG, ",")
	}
	return nil
}
