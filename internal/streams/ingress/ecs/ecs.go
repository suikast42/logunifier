package ecs

import (
	"github.com/nats-io/nats.go"
	"github.com/suikast42/logunifier/internal/streams/ingress"
	"github.com/suikast42/logunifier/pkg/model"
)

type EcsWrapper struct {
	model.EcsLogEntry
}

func (r *EcsWrapper) ConvertToMetaLog(msg *nats.Msg) ingress.IngressMsgContext {
	entry := EcsWrapper{}
	result := ingress.IngressMsgContext{
		NatsMsg: msg,
		MetaLog: &model.MetaLog{
			PatternKey: model.MetaLog_Ecs,
			RawMessage: string(msg.Data),
		},
	}
	err := entry.FromJson(msg.Data)
	r.fillMissing(err, msg, &entry.EcsLogEntry)
	result.MetaLog.EcsLogEntry = &entry.EcsLogEntry
	if err != nil {
		result.MetaLog.EcsLogEntry.ProcessError.Reason = err.Error()
	}
	return result

}

func (r *EcsWrapper) fillMissing(err error, msg *nats.Msg, ecs *model.EcsLogEntry) {
	if ecs.Timestamp == nil {
		ecs.Timestamp = ingress.TimestampFromIngestion(msg)
	}
	if ecs.Log == nil {
		ecs.Log = &model.Log{
			Level: model.LogLevel_not_set,
		}
	}

	ecs.Log.PatternKey = model.MetaLog_Ecs.String()
	ecs.Log.LevelEmoji = model.LogLevelToEmoji(ecs.Log.Level)
	ecs.Log.Ingress = msg.Subject
	if ecs.Id != "" {
		ecs.Id = model.UUID()
	}
	ecs.ProcessError = &model.ProcessError{
		RawData: string(msg.Data),
		Subject: msg.Subject,
	}
	if err != nil {
		ecs.ProcessError.Reason = err.Error()
	}
}
