package converter

import (
	"github.com/nats-io/nats.go"
	"github.com/suikast42/logunifier/pkg/model"
)

type EcsConverter interface {

	// Convert a nats message comes from a Subscription to an EcsLogEntry
	// In case of a marshalling error the converter have fill information in
	// model.ParseError of model.EcsLogEntry
	Convert(msg *nats.Msg) *model.EcsLogEntry
}
