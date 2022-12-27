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

func NewSubscription(name string, durableSubscriptionName string, subscription []string, pushChannel chan<- *ingress.IngressMsgContext) (*ingress.NatsSubscription, error) {
	logger := config.Logger()
	cfg, err := config.Instance()
	if err != nil {
		logger.Error().Err(err).Msgf("Can't obtain config in NewSubscription for  %s", name)
		return nil, err
	}
	//stream cfg
	streamCfg, err := cfg.IngressSubscription(name, "Test ingress for nats cli", subscription)

	if err != nil {
		logger.Error().Err(err).Msgf("Can't create stream config %s", name)
		return nil, err
	}
	return ingress.NewIngresSubscription(durableSubscriptionName, name, subscription, &logger, pushChannel, &TestEcsConverter{}, streamCfg), nil
}

func (r *TestEcsConverter) Convert(msg *nats.Msg) *model.EcsLogEntry {
	r.testCounter++
	logger := config.Logger()
	logger.Debug().Msgf("Counter %d", r.testCounter)
	return &model.EcsLogEntry{
		Id:        model.UUID(),
		Message:   string(msg.Data),
		Timestamp: timestamppb.New(time.Now()),
	}

}
