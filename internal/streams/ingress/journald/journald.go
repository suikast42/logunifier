package journald

import (
	"encoding/json"
	"github.com/nats-io/nats.go"
	"github.com/suikast42/logunifier/internal/config"
	"github.com/suikast42/logunifier/internal/streams/ingress"
	"github.com/suikast42/logunifier/pkg/model"
	"github.com/suikast42/logunifier/pkg/utils"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strconv"
	"strings"
	"time"
)

type JournaldDToEcsConverter struct {
}

// IngressSubjectJournald For journald fields see https://www.freedesktop.org/software/systemd/man/systemd.journal-fields.html
type IngressSubjectJournald struct {
	COM_HASHICORP_NOMAD_ALLOC_ID                  string `json:"COM_HASHICORP_NOMAD_ALLOC_ID"`
	COM_HASHICORP_NOMAD_JOB_ID                    string `json:"COM_HASHICORP_NOMAD_JOB_ID"`
	COM_HASHICORP_NOMAD_JOB_NAME                  string `json:"COM_HASHICORP_NOMAD_JOB_NAME"`
	COM_HASHICORP_NOMAD_NAMESPACE                 string `json:"COM_HASHICORP_NOMAD_NAMESPACE"`
	COM_HASHICORP_NOMAD_NODE_ID                   string `json:"COM_HASHICORP_NOMAD_NODE_ID"`
	COM_HASHICORP_NOMAD_NODE_NAME                 string `json:"COM_HASHICORP_NOMAD_NODE_NAME"`
	COM_HASHICORP_NOMAD_TASK_GROUP_NAME           string `json:"COM_HASHICORP_NOMAD_TASK_GROUP_NAME"`
	COM_HASHICORP_NOMAD_TASK_NAME                 string `json:"COM_HASHICORP_NOMAD_TASK_NAME"`
	COM_GITHUB_LOGUNIFIER_APPLICATION_NAME        string `json:"COM_GITHUB_LOGUNIFIER_APPLICATION_NAME"`
	COM_GITHUB_LOGUNIFIER_APPLICATION_VERSION     string `json:"COM_GITHUB_LOGUNIFIER_APPLICATION_VERSION"`
	COM_GITHUB_LOGUNIFIER_APPLICATION_PATTERN_KEY string `json:"COM_GITHUB_LOGUNIFIER_APPLICATION_PATTERN_KEY"`
	COM_GITHUB_LOGUNIFIER_APPLICATION_STRIP_ANSI  string `json:"COM_GITHUB_LOGUNIFIER_APPLICATION_STRIP_ANSI"`
	CONTAINER_ID                                  string `json:"CONTAINER_ID"`
	CONTAINER_ID_FULL                             string `json:"CONTAINER_ID_FULL"`
	CONTAINER_NAME                                string `json:"CONTAINER_NAME"`
	CONTAINER_TAG                                 string `json:"CONTAINER_TAG"`
	//CONTAINER_PARTIAL_MESSAGE           string    `json:"CONTAINER_PARTIAL_MESSAGE"`
	IMAGE_NAME                        string    `json:"IMAGE_NAME"`
	ORG_OPENCONTAINERS_IMAGE_REVISION string    `json:"ORG_OPENCONTAINERS_IMAGE_REVISION"`
	ORG_OPENCONTAINERS_IMAGE_SOURCE   string    `json:"ORG_OPENCONTAINERS_IMAGE_SOURCE"`
	ORG_OPENCONTAINERS_IMAGE_TITLE    string    `json:"ORG_OPENCONTAINERS_IMAGE_TITLE"`
	PRIORITY                          string    `json:"PRIORITY"`
	SYSLOG_IDENTIFIER                 string    `json:"SYSLOG_IDENTIFIER"`
	SYSLOG_FACILITY                   string    `json:"SYSLOG_FACILITY"`
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
	//if strings.Contains(string(msg.Data), "mimir") {
	//	logger := config.Logger()
	//	logger.Debug().Msg(string(msg.Data))
	//}
	if err != err {
		//logger := config.Logger()
		//logger.Err(err).Msgf("Can't unmarshal journald ingress.\n[%s]", string(msg.Data))
		// The parsing error is shipped to the output
		return ingress.IngressMsgContext{
			NatsMsg: msg,
			MetaLog: &model.MetaLog{
				ApplicationVersion: journald.appVersion(),
				ApplicationName:    journald.appName(),
				PatternKey:         journald.patternKey(),
				ProcessError: &model.ProcessError{
					Reason:  err.Error(),
					RawData: string(msg.Data),
					Subject: msg.Subject,
				},
			},
		}
	}

	return ingress.IngressMsgContext{
		NatsMsg: msg,
		MetaLog: &model.MetaLog{
			ApplicationVersion: journald.appVersion(),
			ApplicationName:    journald.appName(),
			PatternKey:         journald.patternKey(),
			FallbackTimestamp:  journald.ts(),
			FallbackLoglevel:   journald.toLogLevel(),
			Labels:             journald.extractLabels(msg),
			Tags:               journald.tags(),
			Message:            journald.message(),
			ProcessError: &model.ProcessError{
				RawData: string(msg.Data),
				Subject: msg.Subject,
			},
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

	if len(r.COM_HASHICORP_NOMAD_JOB_NAME) > 0 {
		labels[string(model.StaticLabelStack)] = r.COM_HASHICORP_NOMAD_JOB_NAME
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
	if jobType == ingress.JobTypeNomadJob {
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

func (r *IngressSubjectJournald) jobName() string {
	if len(r.COM_HASHICORP_NOMAD_TASK_NAME) > 0 {
		return r.COM_HASHICORP_NOMAD_TASK_NAME
	}
	if len(r.CONTAINER_NAME) > 0 {
		return r.CONTAINER_NAME
	}
	if len(r.SYSTEMDUNIT) > 0 {
		return r.SYSTEMDUNIT
	}

	if len(r.SYSTEMDSLICE) > 0 {
		return r.SYSTEMDSLICE
	}
	if len(r.SYSTEMDCGROUP) > 0 {
		return r.SYSTEMDCGROUP
	}
	if len(r.SYSLOG_IDENTIFIER) > 0 {
		return r.SYSLOG_IDENTIFIER
	}
	// Validation  handles the missing job name
	return ""
}

func (r *IngressSubjectJournald) jobType() ingress.JobType {
	if len(r.COM_HASHICORP_NOMAD_ALLOC_ID) > 0 {
		return ingress.JobTypeNomadJob
	}
	if len(r.CONTAINER_NAME) > 0 {
		return ingress.JobTypeContainer
	}

	switch r.SYSLOG_FACILITY {
	case "0":
		return "kernel"
	case "1":
		return "user"
	case "2":
		return "mail"
	case "3":
		return ingress.JobTypeDaemon
	case "4":
		return "auth"
	case "5":
		return "syslog"
	case "6":
		return "lpr"
	case "7":
		return "news"
	case "8":
		return "uucp"
	case "9":
		return "cron"
	case "10":
		return "authpriv"
	case "11":
		return "ftp"
	case "12":
		return "ntp"
	case "13":
		return "security"
	case "14":
		return "console"
	case "15":
		return "solaris-cron"
	case "16":
		return "local-0"
	case "17":
		return "local-1"
	case "18":
		return "local-2"
	case "19":
		return "local-3"
	case "20":
		return "local-4"
	case "21":
		return "local-5"
	case "22":
		return "local-6"
	case "23":
		return "local-7"

	}

	// Validation  handles the missing job type
	return ""
}

func (r *IngressSubjectJournald) appVersion() string {
	if len(r.COM_GITHUB_LOGUNIFIER_APPLICATION_VERSION) > 0 {
		return r.COM_GITHUB_LOGUNIFIER_APPLICATION_VERSION
	}
	return ""
}

func (r *IngressSubjectJournald) stripAnsi() bool {
	if len(r.COM_GITHUB_LOGUNIFIER_APPLICATION_STRIP_ANSI) > 0 {
		strip, _ := strconv.ParseBool(r.COM_GITHUB_LOGUNIFIER_APPLICATION_STRIP_ANSI)
		return strip
	}
	return false
}

func (r *IngressSubjectJournald) appName() string {
	if len(r.COM_GITHUB_LOGUNIFIER_APPLICATION_NAME) > 0 {
		return r.COM_GITHUB_LOGUNIFIER_APPLICATION_NAME
	}
	return r.jobName()
}

func (r *IngressSubjectJournald) tags() []string {
	if len(r.CONTAINER_TAG) > 0 {
		return strings.Split(r.CONTAINER_TAG, ",")
	}
	return nil
}

func (r *IngressSubjectJournald) patternKey() model.MetaLog_PatternKey {
	if len(r.COM_GITHUB_LOGUNIFIER_APPLICATION_PATTERN_KEY) > 0 {
		return model.StringToLogPatternKey(r.COM_GITHUB_LOGUNIFIER_APPLICATION_PATTERN_KEY)
	}

	return model.MetaLog_Nop
}

func (r *IngressSubjectJournald) message() string {
	if r.stripAnsi() {
		return utils.StripAnsi(r.Message)
	}
	return r.Message
}
