package bootstrap

import (
	"context"
	"github.com/nats-io/nats.go"
	"github.com/suikast42/logunifier/internal/config"
	"github.com/suikast42/logunifier/internal/streams/connectors"
	"github.com/suikast42/logunifier/internal/streams/ingress"
	"github.com/suikast42/logunifier/pkg/model"
	"time"
)

func StreamConfig(streamName string, description string, subjects []string) nats.StreamConfig {
	return nats.StreamConfig{
		Name:         streamName,
		Replicas:     1,
		Description:  description,
		Subjects:     subjects,
		MaxBytes:     1024 * 1024 * 1_000, // 1GB ingress topic
		MaxAge:       time.Hour * 24 * 30, // 30 days
		MaxConsumers: 5,
		Discard:      nats.DiscardOld,
		Retention:    nats.InterestPolicy, //Messages are kept as long as there are Consumers on the stream
		// (matching the message's subject if they are filtered consumers)
		//for which the message has not yet been ACKed.
		//Once all currently defined consumers have received explicit
		//acknowledgement from a subscribing
		//application for the message it is then removed from the stream.
		NoAck:      false,
		Duplicates: time.Minute * 5, // Duplicate time window
		Storage:    nats.FileStorage,
	}
}
func QueueSubscribeConsumerGroupConfig(name string, consumerGroup string, streamConfig nats.StreamConfig, subjectFilter string) nats.ConsumerConfig {
	cfg, _ := config.Instance()
	return nats.ConsumerConfig{
		Durable:       name,
		Name:          name,
		Description:   "Push consumer for subject " + subjectFilter,
		DeliverPolicy: nats.DeliverAllPolicy,
		AckPolicy:     nats.AckExplicitPolicy,
		AckWait:       time.Second * time.Duration(cfg.AckTimeoutS()),
		FilterSubject: subjectFilter,
		MaxAckPending: 1024,
		// That must match with the name for queue subscription
		DeliverGroup:   name,
		DeliverSubject: consumerGroup,
		Replicas:       streamConfig.Replicas,
		MemoryStorage:  false,
		FlowControl:    true,
		Heartbeat:      time.Second * time.Duration(3),
	}
}

func IngressMsgHandler(pushChannel chan<- ingress.IngressMsgContext, metaLogConverter ingress.MetaLogConverter) nats.MsgHandler {
	return func(msg *nats.Msg) {
		pushChannel <- metaLogConverter.ConvertToMetaLog(msg)
	}
}

func EgressMessageHandler(handler connectors.EgressLogHandler) nats.MsgHandler {
	return func(msg *nats.Msg) {
		ecs := &model.EcsLogEntry{}
		err := ecs.FromJson(msg.Data)
		if err != nil {
			logger := config.Logger()
			logger.Error().Err(err).Msgf("Can't deserialize json entry %s", string(msg.Data))
			err := msg.Ack()
			if err != nil {
				logger.Error().Err(err).Msg("Can't ack message")
			}
			return
		}
		handler.Handle(msg, ecs)
	}
}

func ProducerStream(ctx context.Context, conn *nats.Conn) (nats.JetStreamContext, error) {
	return ProducerStreamWithErrorHandler(ctx, conn, nil)
}

func ProducerStreamWithErrorHandler(ctx context.Context, nc *nats.Conn, errorHandler nats.MsgErrHandler) (nats.JetStreamContext, error) {
	// Add at least a default publish error handler even if not defined
	if errorHandler == nil {
		errorHandler = func(stream nats.JetStream, msg *nats.Msg, _err error) {
			logger := config.Logger()
			logger.Error().Err(_err).Msgf("PublishAsyncErrHandler:  Error in send msg to %s", msg.Subject)
		}
	}

	opts := []nats.JSOpt{
		nats.PublishAsyncErrHandler(errorHandler),
		nats.PublishAsyncMaxPending(2 * 1024),
		nats.Context(ctx),
	}

	stream, err := nc.JetStream(opts...)
	if err != nil {
		return nil, err
	}
	return stream, nil
}
