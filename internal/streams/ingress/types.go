package ingress

import (
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
	"github.com/suikast42/logunifier/pkg/model"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strings"
	"time"
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

type JobType string

const (
	JobTypeNomadJob  JobType = "nomad_job"
	JobTypeContainer JobType = "container"
	JobTypeDaemon    JobType = "daemon"
)

func TimestampFromIngestion(msg *nats.Msg) *timestamppb.Timestamp {
	metadata, err := msg.Metadata()
	if err != nil {
		log.Error().Err(err).Msg("Can't extract timestamp from message metadata. Use ts.now() instead")
		return timestamppb.New(time.Now())
	}
	return timestamppb.New(metadata.Timestamp)
}
func HeaderToMap(header nats.Header) map[string]string {
	m := make(map[string]string)
	for k, v := range header {
		m[k] = strings.Join(v, ",")
	}
	return m
}
