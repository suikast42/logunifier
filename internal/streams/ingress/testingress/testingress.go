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

	return ingress.IngressMsgContext{
		NatsMsg: msg,
		MetaLog: &model.MetaLog{
			FallbackTimestamp: timestamppb.New(ts),
			FallbackLoglevel:  model.LogLevel_unknown,
			IsNativeEcs:       false,
			PatternIdentifier: "",
			AppVersion:        "",
			Labels:            extractLabels(msg),
			Message:           string(msg.Data),
			Tags:              nil,
			ParseError:        nil,
		},
	}

}
func extractLabels(msg *nats.Msg) map[string]string {
	var labels = make(map[string]string)
	labels[string(model.StaticLabelIngress)] = msg.Subject
	labels[string(model.StaticLabelJob)] = "test"
	labels[string(model.StaticLabelJobType)] = "test"
	labels[string(model.StaticLabelTask)] = "task"

	return labels
}
