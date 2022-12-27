package ingress

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"github.com/suikast42/logunifier/internal/config"
	"github.com/suikast42/logunifier/pkg/model"
	"strings"
	"sync"
	"time"
)

type EgressLogHandler interface {

	// Handle receives the converted messages from the egress channel and is responsible for
	// sending them to the sink. After the sink accepts the messages the handler is responsible for
	// acknowledging the msg
	Handle(msg *nats.Msg, ecs *model.EcsLogEntry)
}

type EcsConverter interface {

	// Convert a nats message comes from a Subscription to an EcsLogEntry
	// In case of a marshalling error the converter have fill information in
	// model.ParseError of model.EcsLogEntry
	Convert(msg *nats.Msg) *model.EcsLogEntry
}

type IngressMsgContext struct {
	Orig      *nats.Msg
	Converter EcsConverter
}

type IngresSubscriber interface {
	NewSubscription(name string, durableSubscriptionName string, subscription string, pushChannel chan<- *IngressMsgContext) (*NatsSubscription, error)
}
type NatsSubscription struct {
	durableSubscriptionName string
	streamName              string
	subscription            []string
	subscriptionInstance    *nats.Subscription
	logger                  *zerolog.Logger
	ctx                     context.Context
	cancel                  context.CancelFunc
	streamConfig            *nats.StreamConfig
	msgHandler              nats.MsgHandler
	jCtx                    nats.JetStreamContext
}

func (r *NatsSubscription) JCtx() nats.JetStreamContext {
	return r.jCtx
}

func NewIngresSubscription(durableSubscriptionName string,
	streamName string,
	subscription []string,
	logger *zerolog.Logger,
	pushChannel chan<- *IngressMsgContext,
	ecsConverter EcsConverter,
	streamConfig *nats.StreamConfig) *NatsSubscription {

	msgHandler := func(msg *nats.Msg) {
		msgContext := IngressMsgContext{
			Orig:      msg,
			Converter: ecsConverter,
		}
		pushChannel <- &msgContext
	}

	return &NatsSubscription{durableSubscriptionName: durableSubscriptionName,
		streamName:   streamName,
		subscription: subscription,
		logger:       logger,
		streamConfig: streamConfig,
		msgHandler:   msgHandler,
	}
}

func NewEgressSubscription(durableSubscriptionName string,
	streamName string,
	subscription []string,
	logger *zerolog.Logger,
	streamConfig *nats.StreamConfig,
	handler EgressLogHandler,
) *NatsSubscription {

	msgHandler := func(msg *nats.Msg) {
		ecs := &model.EcsLogEntry{}
		err := json.Unmarshal(msg.Data, ecs)
		if err != nil {
			logger.Error().Err(err).Msgf("Can't unmarshal ecs log entry: %s", string(msg.Data))
			err := msg.Ack()
			if err != nil {
				logger.Error().Err(err).Msgf("Can't Ack ecs log entry: %s", string(msg.Data))
			}
			return
		}
		handler.Handle(msg, ecs)
	}

	return &NatsSubscription{durableSubscriptionName: durableSubscriptionName,
		streamName:   streamName,
		subscription: subscription,
		logger:       logger,
		streamConfig: streamConfig,
		msgHandler:   msgHandler,
	}
}

func (r *NatsSubscription) String() string {
	return fmt.Sprintf("%s@%s --> %s", r.durableSubscriptionName, r.streamName, strings.Join(r.subscription, ","))
}

var subscriptionMtx sync.Mutex

func (r *NatsSubscription) Subscribe(ctx context.Context, cancel context.CancelFunc, connection *nats.Conn) error {
	subscriptionMtx.Lock()
	if r.subscriptionInstance != nil {
		r.logger.Error().Msg("Subscription already initialized")
		return nil
	}
	r.ctx = ctx
	r.cancel = cancel
	defer subscriptionMtx.Unlock()
	r.logger.Info().Msgf("Subscribing to %s", r.String())

	f := func(stream nats.JetStream, msg *nats.Msg, _err error) {
		r.logger.Error().Err(_err).Msgf("PublishAsyncErrHandler: %s. Error in send msg", r.streamName)
	}
	opts := []nats.JSOpt{
		nats.PublishAsyncErrHandler(f),
		nats.PublishAsyncMaxPending(2 * 1024),
	}

	js, err := connection.JetStream(opts...)

	if err != nil {
		r.logger.Error().Err(err).Msg("Can't create jetstream connection")
		return err
	}

	cfg, err := config.Instance()
	if err != nil {
		return err
	}
	// stream cfg
	//streamCfg, err := cfg.IngressJournalDConfig(r.streamName)
	//if err != nil {
	//	r.logger.Error().Err(err).Msgf("Can't create stream config %s", r.streamName)
	//	return err
	//}
	err = cfg.CreateOrUpdateStream(r.streamConfig, js)
	if err != nil {
		r.logger.Error().Err(err).Msgf("Can't create stream %s", r.streamName)
		return err
	}

	//// Check if the consumer already exists; if not, create it.
	//consumerInfo, err := js.ConsumerInfo(r.streamName, r.durableSubscriptionName)
	//consumerCfg := &nats.ConsumerConfig{
	//	Durable:       r.durableSubscriptionName,
	//	DeliverPolicy: nats.DeliverAllPolicy,
	//	ReplayPolicy:  nats.ReplayInstantPolicy,
	//	AckPolicy:     nats.AckAllPolicy,
	//	//MaxAckPending: 1000,
	//	//MaxWaiting:    100,
	//	//FlowControl:   true,
	//	//FilterSubject: r.streamName,
	//}
	//if consumerInfo == nil {
	//	consumer, createError := js.AddConsumer(r.streamName, consumerCfg)
	//	if createError != nil {
	//		logger.Error().Err(createError).Msgf("Can't create consumer %s", consumerCfg.Name)
	//		return createError
	//	}
	//	logger.Info().Msgf("Consumer %s created", consumer.Name)
	//} else {
	//	updateConsumer, updateError := js.UpdateConsumer(r.streamName, consumerCfg)
	//	if updateError != nil {
	//		logger.Error().Err(updateError).Msgf("Can't update consumer %s", consumerCfg.Name)
	//		return updateError
	//	}
	//	logger.Info().Msgf("Consumer %s updated", updateConsumer.Name)
	//}
	//subscribe, err := js.PullSubscribe(r.subscription, "Hansi")
	//if err != nil {
	//	logger.Error().Err(err).Stack().Msg("Can't create subscription")
	//	return err
	//}
	subOpts := []nats.SubOpt{
		nats.BindStream(r.streamName),
		nats.Durable(r.durableSubscriptionName),
		nats.AckWait(time.Second * time.Duration(cfg.AckTimeoutS())), // Redeliver after
		//nats.MaxDeliver(5),  // Redeliver max default is infinite
		nats.ManualAck(),         // Control the ack inProgress and nack self
		nats.ReplayInstant(),     // Replay so fast as possible
		nats.DeliverAll(),        // Redeliver all not acked when restarted
		nats.MaxAckPending(1024), // Max inflight ack
		nats.Description(r.streamConfig.Description),
		//nats.EnableFlowControl(),
		//nats.IdleHeartbeat(time.Second * 1),
	}
	subscribe, err := js.QueueSubscribe("", r.streamName, r.msgHandler, subOpts...)
	if err != nil {
		r.logger.Error().Err(err).Msgf("Can't subscribe consumer %s", r.durableSubscriptionName)
		return err
	}
	r.subscriptionInstance = subscribe
	r.jCtx = js
	info, err := subscribe.ConsumerInfo()
	if err != nil {
		r.logger.Error().Err(err).Msgf("Can't obtain consumer info %s", r.durableSubscriptionName)
		return err
	}
	r.logger.Info().Msgf("Subscription %s done", info.Name)
	return nil
}

func (r *NatsSubscription) Unsubscribe() error {
	// Do not use for jet stream subscription.
	// With the subscription created above this unsubscribe will delete the consumer
	// See the doc of r.subscriptionInstance.Unsubscribe()
	//if r.subscriptionInstance != nil {
	//	err := r.subscriptionInstance.Unsubscribe()
	//	if err != nil {
	//		return err
	//	}
	//	r.logger.Info().Msgf("Unsubscribed to %s", r.String())
	//}
	r.subscriptionInstance = nil
	return nil
}
