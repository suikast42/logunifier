package testingress

import (
	"github.com/nats-io/nats.go"
	"github.com/suikast42/logunifier/internal/config"
	"github.com/suikast42/logunifier/internal/streams/ingress"
	"github.com/suikast42/logunifier/pkg/model"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

type TestEcsConverter struct {
	testCounter int64
}

func (r *TestEcsConverter) ConvertToMetaLog(msg *nats.Msg) ingress.IngressMsgContext {
	r.testCounter++
	logger := config.Logger()
	ts, err := time.Parse(time.RFC3339, string(msg.Data))
	if err != nil {
		ts = time.Now()
	}
	logger.Debug().Msgf("Counter %d. Message %s", r.testCounter, string(msg.Data))

	result := ingress.IngressMsgContext{
		NatsMsg: msg,
		MetaLog: &model.MetaLog{
			PatternKey: model.MetaLog_Ecs,
			RawMessage: string(msg.Data),
			EcsLogEntry: &model.EcsLogEntry{
				Timestamp: timestamppb.New(ts),
				Message:   string(msg.Data),
			},
		},
	}
	r.extractServiceMetadata(result.MetaLog.EcsLogEntry)
	return result

}
func (r *TestEcsConverter) extractServiceMetadata(ecs *model.EcsLogEntry) {
	if ecs.Service == nil {
		ecs.Service = &model.Service{}
	}
	ecs.Service.Stack = "Test"
	ecs.Service.Group = "Test"
	ecs.Service.Namespace = "Test"
	ecs.Service.Name = "Test"
}
