package config

import (
	"errors"
	"github.com/nats-io/nats.go"
	"time"
)

func (c Config) StreamConfig(streamName string, description string, subjects []string) (*nats.StreamConfig, error) {
	streamCfg := &nats.StreamConfig{
		Name:         streamName,
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
	return streamCfg, nil
}

func (c Config) CreateOrUpdateStream(streamcfg *nats.StreamConfig, js nats.JetStreamContext) error {

	// Check if the stream already exists; if not, create it.
	logger := Logger()
	streamInfo, err := js.StreamInfo(streamcfg.Name)
	if err != nil {
		apiErr := &nats.APIError{}
		if !errors.As(err, &apiErr) || apiErr.ErrorCode != nats.JSErrCodeStreamNotFound {
			return err
		}
	}

	if streamInfo == nil {
		// Create a stream
		stream, err := js.AddStream(streamcfg)
		if err != nil {
			apiErr := &nats.APIError{}
			if errors.As(err, &apiErr) {
				logger.Error().Err(apiErr).Msgf("Can't add stream %s. ErrorCode %v. Code: %v. Description: %v  ", streamcfg.Name, apiErr.ErrorCode, apiErr.Code, apiErr.Description)
			} else {
				logger.Error().Err(err).Msgf("Can't add stream %s", streamcfg.Name)
			}
			return err
		}
		logger.Info().Msgf("Connected to stream streamName: %s", stream.Config.Name)
	} else {
		// Update a stream
		updateStream, err := js.UpdateStream(streamcfg)
		if err != nil {
			logger.Error().Err(err).Msgf("Can't update stream %s", streamcfg.Name)
			return err
		}
		logger.Info().Msgf("Updated to stream streamName: %s", updateStream.Config.Name)
	}
	return nil
}
