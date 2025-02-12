package config

import (
	"errors"
	"github.com/nats-io/nats.go"
)

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
