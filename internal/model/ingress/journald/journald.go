package journald

import (
	"context"
	"fmt"
	"github.com/nats-io/nats.go"
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
	name         string
	subscription string
}

func NewSubscription(name string, subscription string) *IngressJournaldSubscription {
	return &IngressJournaldSubscription{
		name:         name,
		subscription: subscription,
	}
}

func (r *IngressJournaldSubscription) String() string {
	return fmt.Sprintf("%s --> %s", r.name, r.subscription)
}

func (r *IngressJournaldSubscription) Subscribe(ctx context.Context, cancel context.CancelFunc, connection *nats.Conn) error {
	logger := config.Logger()
	logger.Info().Msgf("Subscribing to %s", r.String())

	js, err := connection.JetStream()

	if err != nil {
		logger.Error().Err(err).Msg("Can't create jetstream connection")
		return err
	}

	// stream cfg
	streamcfg := &nats.StreamConfig{
		Name:         r.name,
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
	// Create a stream
	stream, err := js.AddStream(streamcfg)
	if err != nil {
		logger.Error().Err(err).Msgf("Can't add stream %s", streamcfg.Name)
		return err
	}
	logger.Info().Msgf("Connected to stream name: %s", stream.Config.Name)
	// Update a stream
	updateStream, err := js.UpdateStream(streamcfg)
	if err != nil {
		logger.Error().Err(err).Msgf("Can't update stream %s", streamcfg.Name)
		return err
	}
	logger.Info().Msgf("Updated to stream name: %s", updateStream.Config.Name)
	// Create a Consumer
	js.AddConsumer(r.name,
		&nats.ConsumerConfig{
			Durable: "JournaldIngressProcessor",
		})

	return nil
}
