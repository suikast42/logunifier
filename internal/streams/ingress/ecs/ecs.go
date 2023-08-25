package ecs

import (
	"github.com/nats-io/nats.go"
	"github.com/suikast42/logunifier/internal/streams/ingress"
	"github.com/suikast42/logunifier/pkg/model"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type EcsWrapper struct {
	model.EcsLogEntry
}

func (r *EcsWrapper) ConvertToMetaLog(msg *nats.Msg) ingress.IngressMsgContext {
	entry := EcsWrapper{}
	err := entry.FromJson(msg.Data)

	if err != nil {
		//logger := config.Logger()
		//logger.Err(err).Msgf("Can't unmarshal journald ingress.\n[%s]", string(msg.Data))
		// The parsing error is shipped to the output
		return ingress.IngressMsgContext{
			NatsMsg: msg,

			MetaLog: &model.MetaLog{
				ApplicationVersion: "",
				ApplicationName:    "",
				Labels:             extractLabels(msg),
				PatternKey:         model.MetaLog_Ecs,
				FallbackTimestamp:  timestamppb.Now(),
				FallbackLoglevel:   model.LogLevel_error,
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
			Message:    string(msg.Data),
			Labels:     extractLabels(msg),
			PatternKey: model.MetaLog_Ecs,
			ProcessError: &model.ProcessError{
				RawData: string(msg.Data),
				Subject: msg.Subject,
			},
		},
	}
}

func extractLabels(msg *nats.Msg) map[string]string {
	var labels = make(map[string]string)
	labels[string(model.StaticLabelIngress)] = msg.Subject

	return labels
}
