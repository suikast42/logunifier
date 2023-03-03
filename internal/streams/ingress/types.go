package ingress

import (
	"github.com/nats-io/nats.go"
	"github.com/suikast42/logunifier/pkg/model"
	"strings"
)

type MetaLogConverter interface {

	// ConvertToMetaLog a nats message comes from a Subscription to model.MetaLog
	// In case of a marshalling error the converter have fill information in
	// model.MetaLog#ParseError of model.EcsLogEntry
	ConvertToMetaLog(msg *nats.Msg) IngressMsgContext
}

type IngressMsgContext struct {
	NatsMsg *nats.Msg
	MetaLog *model.MetaLog
}

// LabelStatic. Labels can be emmited during ingress phase

var JobTypes = []JobType{
	JobTypeOsService,
	JobTypeNomadJob,
	JobTypeContainer,
}

type JobType string

const (
	JobTypeOsService JobType = "os_service"
	JobTypeNomadJob  JobType = "nomad_service"
	JobTypeContainer JobType = "container"
)

func HeaderToMap(header nats.Header) map[string]string {
	m := make(map[string]string)
	for k, v := range header {
		m[k] = strings.Join(v, ",")
	}
	return m
}
