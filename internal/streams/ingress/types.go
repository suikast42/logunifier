package ingress

import (
	"context"
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"github.com/suikast42/logunifier/internal/config"
	"github.com/suikast42/logunifier/internal/streams/converter"
	"github.com/suikast42/logunifier/internal/streams/egress"
	"sync"
	"time"
)

type IngresSubscriber interface {
	NewSubscription(name string, durableSubscriptionName string, subscription string, pushChannel chan<- *egress.MsgContext) (*IngresSubscription, error)
}
type IngresSubscription struct {
	durableSubscriptionName string
	streamName              string
	subscription            string
	subscriptionInstance    *nats.Subscription
	logger                  *zerolog.Logger
	ctx                     context.Context
	cancel                  context.CancelFunc
	pushChannel             chan<- *egress.MsgContext
	ecsConverter            converter.EcsConverter
	streamConfig            *nats.StreamConfig
}

func NewIngresSubscription(durableSubscriptionName string,
	streamName string,
	subscription string,
	logger *zerolog.Logger,
	pushChannel chan<- *egress.MsgContext,
	ecsConverter converter.EcsConverter,
	streamConfig *nats.StreamConfig) *IngresSubscription {
	return &IngresSubscription{durableSubscriptionName: durableSubscriptionName,
		streamName:   streamName,
		subscription: subscription,
		logger:       logger, pushChannel: pushChannel,
		ecsConverter: ecsConverter,
		streamConfig: streamConfig}
}

func (r *IngresSubscription) String() string {
	return fmt.Sprintf("%s@%s --> %s", r.durableSubscriptionName, r.streamName, r.subscription)
}

var subscriptionMtx sync.Mutex

func (r *IngresSubscription) Subscribe(ctx context.Context, cancel context.CancelFunc, connection *nats.Conn) error {
	subscriptionMtx.Lock()
	if r.subscriptionInstance != nil {
		r.logger.Error().Msg("Subscription already initialized")
		return nil
	}
	r.ctx = ctx
	r.cancel = cancel
	defer subscriptionMtx.Unlock()
	r.logger.Info().Msgf("Subscribing to %s", r.String())

	js, err := connection.JetStream()

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
		nats.AckWait(time.Second * time.Duration(cfg.AckTimeoutS())), // Redeliver after
		//nats.MaxDeliver(5),  // Redeliver max default is infinite
		nats.ManualAck(),         // Control the ack inProgress and nack self
		nats.ReplayInstant(),     // Replay so fast as possible
		nats.DeliverAll(),        // Redeliver all not acked when restarted
		nats.MaxAckPending(1024), // Max inflight ack
		nats.EnableFlowControl(),
		nats.IdleHeartbeat(time.Second * 1),
		nats.Durable(r.durableSubscriptionName),
	}
	subscribe, err := js.Subscribe("", func(msg *nats.Msg) {
		msgContext := egress.MsgContext{
			Orig:      msg,
			Converter: r.ecsConverter,
		}
		r.pushChannel <- &msgContext
		//msg.InProgress()
		//journald := IngressSubjectJournald{}
		//unmarsahlError := json.Unmarshal(msg.Data, &journald)
		//if unmarsahlError != nil {
		//	r.logger.Error().Err(unmarsahlError).Msg("Can't unmarshal message fom journald channel")
		//	err := msg.Term()
		//	if err != nil {
		//		return
		//	}
		//}
		//r.logger.Info().Msgf("Received message %s - %s", journald.Timestamp, journald.Message)
		//msg.Ack()
	}, subOpts...)
	if err != nil {
		r.logger.Error().Err(err).Msgf("Can't subscribe consumer %s", r.durableSubscriptionName)
		return err
	}
	r.subscriptionInstance = subscribe
	info, err := subscribe.ConsumerInfo()
	if err != nil {
		r.logger.Error().Err(err).Msgf("Can't obtain consumer info %s", r.durableSubscriptionName)
		return err
	}
	r.logger.Info().Msgf("Subscription %s done", info.Name)
	return nil
}

func (r *IngresSubscription) Unsubscribe() error {
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
	return nil
}
