package connectors

import (
	"github.com/nats-io/nats.go"
	"github.com/suikast42/logunifier/pkg/model"
)

type EgressLogHandler interface {

	// Handle receives the converted messages from the egress channel and is responsible for
	// sending them to the sink. After the sink accepts the messages the handler is responsible for
	// acknowledging the msg
	Handle(msg *nats.Msg, ecs *model.EcsLogEntry)
}
