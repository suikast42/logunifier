package lokishipper

import (
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"github.com/suikast42/logunifier/pkg/model"
	"time"
)

type LokiShipper struct {
	//GRPC_GO_LOG_VERBOSITY_LEVEL=99
	//GRPC_GO_LOG_SEVERITY_LEVEL=info
	Logger        zerolog.Logger
	LokiAddresses []string
	connected     bool
}

func (loki *LokiShipper) Handle(msg *nats.Msg, ecs *model.EcsLogEntry) {
	if !loki.connected {
		loki.Logger.Error().Msg("not connected to loki")
		err := msg.NakWithDelay(time.Second * 5)
		if err != nil {
			loki.Logger.Error().Err(err).Msg("Can't nack message")
		}
	}
	//loki.Logger.Info().Msgf("LokiShipper Received message: %s", ecs.Message)
	err := msg.Ack()
	if err != nil {
		loki.Logger.Err(err).Msg("Can't ack message")
	}
}

func (loki *LokiShipper) Connect() {

}

func (loki *LokiShipper) DisConnect() {

}
