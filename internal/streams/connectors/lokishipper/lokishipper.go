package lokishipper

import (
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"github.com/suikast42/logunifier/pkg/model"
)

type LokiShipper struct {
	Logger zerolog.Logger
}

func (loki *LokiShipper) Handle(msg *nats.Msg, ecs *model.EcsLogEntry) {
	loki.Logger.Info().Msgf("LokiShipper Received message: %s", ecs.Message)
	err := msg.Ack()
	if err != nil {
		loki.Logger.Err(err).Msg("Can't ack message")
	}
}
