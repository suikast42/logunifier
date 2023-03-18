package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"github.com/suikast42/logunifier/internal/config"
	"github.com/suikast42/logunifier/pkg/utils"
	"os"
	"sync"
	"time"
)
import _ "github.com/nats-io/nats.go"

// NatsDialer Wraps the context of subscriptions and nats server connection
type NatsDialer struct {
	ctx               context.Context
	logger            *zerolog.Logger
	nc                *nats.Conn
	connectionTimeOut time.Duration
	connectTimeWait   time.Duration
	// Streams to create or update after the connection is done to nats server
	// The key should be equal to the stream name
	streamConfigurations map[string]NatsStreamConfiguration
	// The key should be the same as the Consumer name
	consumerConfigurations map[string]NatsConsumerConfiguration
}

// NatsStreamConfiguration definition of a nats stream
type NatsStreamConfiguration struct {
	// The Nats Stream configuration
	StreamConfiguration nats.StreamConfig
}

type NatsConsumerConfiguration struct {
	// The consumer configuration
	ConsumerConfiguration nats.ConsumerConfig

	// That name must match with a defined stream cfg in streamConfigurations of NatsDialer
	StreamName string

	// Handler that receives the incoming message.
	// The handler is responsible for ack and nack the message
	MsgHandler nats.MsgHandler
	//for internal usage

}

var instance *NatsDialer
var instanceLock sync.Mutex

func New(streamConfigurations map[string]NatsStreamConfiguration, consumerConfigurations map[string]NatsConsumerConfiguration) (*NatsDialer, error) {
	instanceLock.Lock()
	defer instanceLock.Unlock()
	if instance == nil {
		logger := config.Logger()
		instance = &NatsDialer{
			ctx:                    context.Background(),
			logger:                 &logger,
			connectionTimeOut:      time.Second * 1,
			connectTimeWait:        time.Second * 1,
			streamConfigurations:   streamConfigurations,
			consumerConfigurations: consumerConfigurations,
		}
		return instance, nil
	}
	return nil, errors.New("already initialized. Use Intance() instead")
}

func Intance() (*NatsDialer, error) {
	instanceLock.Lock()
	defer instanceLock.Unlock()
	if instance == nil {
		return nil, errors.New("not initialized. Use New() at first")
	}
	return instance, nil
}

var connectionMtx sync.Mutex

func (nd *NatsDialer) Connection() *nats.Conn {
	return nd.nc
}

func (nd *NatsDialer) Health(ctx context.Context) error {
	if nd.nc == nil {
		return errors.New("connection is not set")
	}

	if !nd.nc.IsConnected() {
		return errors.New("not connected yet")
	}
	return nil
}
func (nd *NatsDialer) Connect() error {
	connectionMtx.Lock()
	cfg, err := config.Instance()
	if err != nil {
		return err
	}
	//if nd.nc != nil {
	//	//Connection already established or establishing
	//	return nil
	//}
	defer connectionMtx.Unlock()

	opts := []nats.Option{
		nats.Name("logunifier"),
		nats.Timeout(nd.connectionTimeOut),
		nats.RetryOnFailedConnect(true),
		nats.ConnectHandler(func(nc *nats.Conn) {
			nd.logger.Info().Msgf("Connected to  %s", nc.ConnectedUrl())
			nd.nc = nc
			nd.doSubscribe()
		}),
		nats.ClosedHandler(func(c *nats.Conn) {
			closeByRequest := nd.ctx.Value("DisconnectRequest")
			if closeByRequest != nil && closeByRequest.(bool) {
				// This is the case if the connection loosed by the program itself
				nd.logger.Info().Msgf("Connection closed to %s by a DisconnectRequest", c.ConnectedUrl())
			} else {
				// Fatal do a os.Exit(1)
				nd.logger.Fatal().Msgf("Can't connect to %s connection lost", c.ConnectedUrl())
				//
				//// This is the case if the nats clients lost the connection
				//nd.logger.Warn().Msgf("Connection to %s lost. Reconnect", c.ConnectedUrl())
				//go func() {
				//	connErr := nd.Connect()
				//	if connErr != nil {
				//		// Fatal do a os.Exit(1)
				//		nd.logger.Fatal().Msgf("Can't connect to %s after connection lost")
				//	}
				//}()
			}

		}),
		nats.ReconnectWait(nd.connectTimeWait),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			nd.logger.Info().Msgf("Reconnected to %s", nc.ConnectedUrl())
			nd.nc = nc
			nd.doSubscribe()
		}),
		nats.DisconnectErrHandler(func(c *nats.Conn, disconnectionError error) {
			if disconnectionError != nil {
				nd.logger.Error().Err(disconnectionError).Msg("Disconnection with error")
			} else {
				nd.logger.Info().Msgf("Disconnected from NATS %s", c.ConnectedUrl())
			}
		}),

		//This will kill the client if the connection is lost
		//We will keep the connection
		//nats.NoReconnect(),
	}
	_, err = nats.Connect(cfg.NatsServers(), opts...)
	if err != nil {
		return err
	}

	return nil
}

func (nd *NatsDialer) Disconnect() error {
	if nd.nc == nil || !nd.nc.IsConnected() {
		nd.logger.Info().Msg("Not connected to server nothing todo")
		return nil
	}
	// Disconnect and flush pending messages
	if err := nd.nc.Drain(); err != nil {
		nd.logger.Error().Err(err).Msg("Can't Drain")
		return err
	}
	nd.ctx = context.WithValue(nd.ctx, "DisconnectRequest", true)
	nd.nc.Close()
	nd.logger.Info().Msg("Disconnected")
	return nil
}
func (nd *NatsDialer) doSubscribe() {
	streamDefinitionError := nd.upsertStreams()
	if streamDefinitionError != nil {
		nd.logger.Error().Err(streamDefinitionError).Msgf("Can't create or update stream(s) for config %+v", nd.streamConfigurations)
		os.Exit(1)
	}

	consumerDefinitionError := nd.upsertConsumers()
	if consumerDefinitionError != nil {
		nd.logger.Error().Err(consumerDefinitionError).Msgf("Can't create or update consumer(s) for config %+v", nd.consumerConfigurations)
		os.Exit(1)
	}

	subscriptionError := nd.startSubscriptions()
	if subscriptionError != nil {
		nd.logger.Error().Err(subscriptionError).Msgf("Can't start subscription %+v", nd.consumerConfigurations)
		os.Exit(1)
	}
}

func (nd *NatsDialer) upsertStreams() error {
	logger := nd.logger
	for key, definition := range nd.streamConfigurations {
		logger.Info().Msgf("UpsertStream for definition key %s", key)

		js, err := nd.nc.JetStream()
		if err != nil {
			if err != nil {
				logger.Error().Err(err).Msgf("Can't create JetStream for %+v", definition)
				return err
			}
		}
		streamInfo, err := js.StreamInfo(definition.StreamConfiguration.Name)
		if err != nil {
			apiErr := &nats.APIError{}
			if !errors.As(err, &apiErr) || apiErr.ErrorCode != nats.JSErrCodeStreamNotFound {
				return err
			}
		}

		if streamInfo == nil {
			// Create a stream
			streamInfo, err := js.AddStream(&definition.StreamConfiguration)
			if err != nil {
				apiErr := &nats.APIError{}
				if errors.As(err, &apiErr) {
					logger.Error().Err(apiErr).Msgf("Can't add stream %s. ErrorCode %v. Code: %v. Description: %v  ", definition.StreamConfiguration.Name, apiErr.ErrorCode, apiErr.Code, apiErr.Description)
				} else {
					logger.Error().Err(err).Msgf("Can't add stream %s", definition.StreamConfiguration.Name)
				}
				return err
			}
			logger.Info().Msgf("Connected to stream streamName: %s", streamInfo.Config.Name)
		} else {
			// Update a stream
			updateStreamInfo, err := js.UpdateStream(&definition.StreamConfiguration)
			if err != nil {
				logger.Error().Err(err).Msgf("Can't update stream %s", definition.StreamConfiguration.Name)
				return err
			}
			logger.Info().Msgf("Updated to stream streamName: %s", updateStreamInfo.Config.Name)
		}
	}

	return nil
}

func (nd *NatsDialer) upsertConsumers() error {
	logger := nd.logger
	for key, definition := range nd.consumerConfigurations {
		logger.Info().Msgf("UpsertConsumer for definition key %s", key)
		//js := nd.streamConfigurations[definition.StreamName].streamCtx
		//if js == nil {
		//	return errors.New(fmt.Sprintf("No JestStreamCtx found for consumer %s for stream  %s", definition.ConsumerConfiguration.Name, definition.StreamName))
		//}
		js, err := nd.nc.JetStream()
		if err != nil {
			if err != nil {
				logger.Error().Err(err).Msgf("Can't create JetStream for %+v", definition)
				return err
			}
		}
		consumerInfo, consumerInfoError := js.ConsumerInfo(definition.StreamName, definition.ConsumerConfiguration.Name)
		if consumerInfoError != nil {
			apiErr := &nats.APIError{}
			if !errors.As(consumerInfoError, &apiErr) || apiErr.ErrorCode != nats.JSErrCodeConsumerNotFound {
				return consumerInfoError
			}
		}
		if consumerInfo == nil {
			consumer, err := js.AddConsumer(definition.StreamName, &definition.ConsumerConfiguration)
			if err != nil {
				apiErr := &nats.APIError{}
				if errors.As(err, &apiErr) {
					logger.Error().Err(apiErr).Msgf("Can't add consumer %s. ErrorCode %v. Code: %v. Description: %v  ", definition.ConsumerConfiguration.Name, apiErr.ErrorCode, apiErr.Code, apiErr.Description)
				} else {
					logger.Error().Err(err).Msgf("Can't add consumer %s", definition.ConsumerConfiguration.Name)
				}
				return err
			}
			logger.Info().Msgf("Consumer %s added to stream streamName: %s", consumer.Name, consumer.Stream)
		} else {
			consumer, err := js.UpdateConsumer(definition.StreamName, &definition.ConsumerConfiguration)
			if err != nil {
				apiErr := &nats.APIError{}
				if errors.As(err, &apiErr) {
					logger.Error().Err(apiErr).Msgf("Can't update consumer %s. ErrorCode %v. Code: %v. Description: %v  ", definition.ConsumerConfiguration.Name, apiErr.ErrorCode, apiErr.Code, apiErr.Description)
				} else {
					logger.Error().Err(err).Msgf("Can't update consumer %s", definition.ConsumerConfiguration.Name)
				}
				return err
			}
			logger.Info().Msgf("Consumer %s updated for stream streamName: %s", consumer.Name, consumer.Stream)
		}
	}
	return nil
}

func (nd *NatsDialer) startSubscriptions() error {
	logger := nd.logger
	//js := configuration.streamCtx
	js, err := nd.nc.JetStream()
	if err != nil {
		if err != nil {
			logger.Error().Err(err).Msg("startSubscriptions: Can't create JetStream for ")
			return err
		}
	}
	for key, definition := range nd.consumerConfigurations {
		logger.Info().Msgf("Start subscription for consumer %s at stream %s and subject %s", key, definition.StreamName, definition.ConsumerConfiguration.FilterSubject)
		_, exists := nd.streamConfigurations[definition.StreamName]
		if !exists {
			return errors.New(fmt.Sprintf("No stream configuration found for client configuration %+v ", definition))
		}

		subOpts := []nats.SubOpt{
			nats.BindStream(definition.StreamName),
			nats.Durable(definition.ConsumerConfiguration.Durable),
			nats.ManualAck(), // Control the ack inProgress and nack self
		}
		_, err = js.QueueSubscribe(definition.ConsumerConfiguration.FilterSubject, definition.ConsumerConfiguration.DeliverGroup, definition.MsgHandler, subOpts...)
		if err != nil {
			logger.Error().Err(err).Msgf("QueueSubscribe to %s failed", definition.ConsumerConfiguration.Name)
			return err
		}
	}

	for streamInfo := range js.Streams() {
		consumers := js.Consumers(streamInfo.Config.Name)
		var consumerInfos []*nats.ConsumerInfo
		for value := range consumers {
			consumerInfos = append(consumerInfos, value)
		}
		buffer := utils.NewStringBuffer()
		buffer.AppendFormat("Created consumers for stream %s", streamInfo.Config.Name)
		buffer.AppendLine("")
		for i, consumerInfo := range consumerInfos {
			buffer.AppendFormat("Name: %s, Durable: %s, FilterSubject:%s ,DeliverGroup: %s ", consumerInfo.Name, consumerInfo.Config.Durable, consumerInfo.Config.FilterSubject, consumerInfo.Config.DeliverGroup)
			if i != len(consumerInfos)-1 {
				buffer.AppendLine("")
			}
		}
		logger.Info().Msg(buffer.ToString())
	}
	return nil
}
func (nd *NatsDialer) SendPing() error {
	if nd.nc == nil || !nd.nc.IsConnected() {
		return errors.New("not connected to server")
	}
	err := nd.nc.Publish("ping", []byte("ping"))
	if err != nil {
		return err
	}
	return nil
}
