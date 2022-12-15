package journald

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"github.com/suikast42/logunifier/internal/config"
	"time"
)

type IngressSubjectJournald struct {
	PRIORITY            string    `json:"PRIORITY"`
	SYSLOGFACILITY      string    `json:"SYSLOG_FACILITY"`
	SYSLOGIDENTIFIER    string    `json:"SYSLOG_IDENTIFIER"`
	BOOTID              string    `json:"_BOOT_ID"`
	CAPEFFECTIVE        string    `json:"_CAP_EFFECTIVE"`
	CMDLINE             string    `json:"_CMDLINE"`
	COMM                string    `json:"_COMM"`
	EXE                 string    `json:"_EXE"`
	GID                 string    `json:"_GID"`
	MACHINEID           string    `json:"_MACHINE_ID"`
	PID                 string    `json:"_PID"`
	SELINUXCONTEXT      string    `json:"_SELINUX_CONTEXT"`
	STREAMID            string    `json:"_STREAM_ID"`
	SYSTEMDCGROUP       string    `json:"_SYSTEMD_CGROUP"`
	SYSTEMDINVOCATIONID string    `json:"_SYSTEMD_INVOCATION_ID"`
	SYSTEMDSLICE        string    `json:"_SYSTEMD_SLICE"`
	SYSTEMDUNIT         string    `json:"_SYSTEMD_UNIT"`
	TRANSPORT           string    `json:"_TRANSPORT"`
	UID                 string    `json:"_UID"`
	MONOTONICTIMESTAMP  string    `json:"__MONOTONIC_TIMESTAMP"`
	REALTIMETIMESTAMP   string    `json:"__REALTIME_TIMESTAMP"`
	Host                string    `json:"host"`
	Message             string    `json:"message"`
	SourceType          string    `json:"source_type"`
	Timestamp           time.Time `json:"timestamp"`
}

type IngressJournaldSubscription struct {
	durableSubscriptionName string
	streamName              string
	subscription            string
	subscriptionInstance    *nats.Subscription
	logger                  *zerolog.Logger
}

func NewSubscription(name string, durableSubscriptionName string, subscription string) *IngressJournaldSubscription {
	logger := config.Logger()
	return &IngressJournaldSubscription{
		durableSubscriptionName: durableSubscriptionName,
		streamName:              name,
		subscription:            subscription,
		logger:                  &logger,
	}
}

func (r *IngressJournaldSubscription) String() string {
	return fmt.Sprintf("%s@%s --> %s", r.durableSubscriptionName, r.streamName, r.subscription)
}

func (r *IngressJournaldSubscription) Subscribe(ctx context.Context, cancel context.CancelFunc, connection *nats.Conn) error {
	r.logger.Info().Msgf("Subscribing to %s", r.String())

	js, err := connection.JetStream()

	if err != nil {
		r.logger.Error().Err(err).Msg("Can't create jetstream connection")
		return err
	}

	// stream cfg
	streamcfg := &nats.StreamConfig{
		Name:         r.streamName,
		Description:  "Ingress Processor for journald logs comes over vector",
		Subjects:     []string{r.subscription},
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

	}
	// Check if the stream already exists; if not, create it.
	streamInfo, err := js.StreamInfo(streamcfg.Name)

	if streamInfo == nil {
		// Create a stream
		stream, err := js.AddStream(streamcfg)
		if err != nil {
			r.logger.Error().Err(err).Msgf("Can't add stream %s", streamcfg.Name)
			return err
		}
		r.logger.Info().Msgf("Connected to stream streamName: %s", stream.Config.Name)
	} else {
		// Update a stream
		updateStream, err := js.UpdateStream(streamcfg)
		if err != nil {
			r.logger.Error().Err(err).Msgf("Can't update stream %s", streamcfg.Name)
			return err
		}
		r.logger.Info().Msgf("Updated to stream streamName: %s", updateStream.Config.Name)
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
		nats.AckWait(time.Second * 10), // Redeliver after
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
		msg.InProgress()
		journald := IngressSubjectJournald{}
		unmarsahlError := json.Unmarshal(msg.Data, &journald)
		if unmarsahlError != nil {
			r.logger.Error().Err(unmarsahlError).Msg("Can't unmarshal message fom journald channel")
			err := msg.Term()
			if err != nil {
				return
			}
		}
		r.logger.Info().Msgf("Received message %s - %s", journald.Timestamp, journald.Message)
		msg.Ack()
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

func (r *IngressJournaldSubscription) Unsubscribe() error {
	if r.subscriptionInstance != nil {
		err := r.subscriptionInstance.Unsubscribe()
		if err != nil {
			return err
		}
		r.logger.Info().Msgf("Unsubscribed to %s", r.String())
	}
	return nil
}

func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}
