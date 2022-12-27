package lokishipper

import (
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"github.com/suikast42/logunifier/internal/config"
	"github.com/suikast42/logunifier/internal/streams/ingress"
	"github.com/suikast42/logunifier/pkg/model"
)

type LokiShipper struct {
	logger *zerolog.Logger
}

func NewSubscription(name string, durableSubscriptionName string, subscription []string) (*ingress.NatsSubscription, error) {
	logger := config.Logger()
	cfg, err := config.Instance()
	if err != nil {
		logger.Error().Err(err).Msgf("Can't obtain config in NewSubscription for  %s", name)
		return nil, err
	}

	//stream cfg
	streamCfg, err := cfg.StreamConfig(name, "Egress channel for ecs logs", subscription)

	if err != nil {
		logger.Error().Err(err).Msgf("Can't create stream config %s", name)
		return nil, err
	}
	return ingress.NewEgressSubscription(durableSubscriptionName, name, subscription, &logger, streamCfg, &LokiShipper{
		logger: &logger,
	}), nil
}

func (loki *LokiShipper) Handle(msg *nats.Msg, ecs *model.EcsLogEntry) {
	loki.logger.Info().Msgf("Received message: %s", ecs.Message)
	err := msg.Ack()
	if err != nil {
		loki.logger.Err(err).Msg("Can't ack message")
	}
}
