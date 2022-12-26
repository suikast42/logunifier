package testingress

import (
	"github.com/nats-io/nats.go"
	"github.com/suikast42/logunifier/internal/config"
	"github.com/suikast42/logunifier/internal/streams/egress"
	"github.com/suikast42/logunifier/internal/streams/ingress"
	"github.com/suikast42/logunifier/pkg/model"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

type TestEcsConverter struct {
}

func NewSubscription(name string, durableSubscriptionName string, subscription string, pushChannel chan<- *egress.MsgContext) *ingress.IngresSubscription {
	logger := config.Logger()
	return ingress.NewIngresSubscription(durableSubscriptionName, name, subscription, &logger, pushChannel, &TestEcsConverter{})
}

func (r *TestEcsConverter) Convert(msg *nats.Msg) *model.EcsLogEntry {
	return &model.EcsLogEntry{
		Id:        "1",
		Message:   string(msg.Data),
		Timestamp: timestamppb.New(time.Now()),
	}

}
