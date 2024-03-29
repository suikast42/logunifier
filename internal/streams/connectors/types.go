package connectors

import (
	"github.com/nats-io/nats.go"
	"github.com/suikast42/logunifier/pkg/model"
)

type EgressMsgContext struct {
	NatsMsg *nats.Msg
	Ecs     *model.EcsLogEntry
}

type EgressLogHandler interface {

	// Handle receives the converted messages from the egress channel and is responsible for
	// sending them to the sink. After the sink accepts the messages the handler is responsible for
	// acknowledging the msg
	Handle(msg *nats.Msg, ecs *model.EcsLogEntry)
}

type ExternalConnection interface {
	// Connect Emit the connection request
	Connect()

	// DisConnect if the connection is established
	DisConnect()
}
