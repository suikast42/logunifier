package ingress

import (
	"github.com/nats-io/nats.go"
	"github.com/suikast42/logunifier/pkg/model"
)

type EcsConverter interface {

	// Convert a nats message comes from a Subscription to an EcsLogEntry
	// In case of a marshalling error the converter have fill information in
	// model.ParseError of model.EcsLogEntry
	// This step have to add following labels:
	// 		ingress: the semantic name of ingress stream like journald, container
	// 		used_grok: the used pattern for that EcsLogEntry see patterns.PatternKey
	Convert(msg *nats.Msg) *model.EcsLogEntry
}

type IngressMsgContext struct {
	Orig      *nats.Msg
	Converter EcsConverter
}
