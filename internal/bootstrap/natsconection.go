package bootstrap

import (
	"context"
	"errors"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"github.com/suikast42/logunifier/internal/config"
	"sync"
	"time"
)
import _ "github.com/nats-io/nats.go"

// NatsDialer Wraps the context of subscriptions and nats server connection
type NatsDialer struct {
	ctx               context.Context
	logger            *zerolog.Logger
	nc                *nats.Conn
	subscriptions     []NatsSubscription
	connectionTimeOut time.Duration
	connectTimeWait   time.Duration
}

func New(subscriptions []NatsSubscription) *NatsDialer {
	logger := config.Logger()
	return &NatsDialer{
		ctx:               context.Background(),
		subscriptions:     subscriptions,
		logger:            &logger,
		connectionTimeOut: time.Second * 1,
		connectTimeWait:   time.Second * 1,
	}
}

var connectionMtx sync.Mutex

func (nd *NatsDialer) Connection() *nats.Conn {
	return nd.nc
}
func (nd *NatsDialer) Connect() error {
	connectionMtx.Lock()
	cfg, err := config.Instance()
	if err != nil {
		return err
	}
	if nd.nc != nil {
		//Connection already established or establishing
		return nil
	}
	defer connectionMtx.Unlock()

	opts := []nats.Option{
		nats.Name("logunifier"),
		nats.Timeout(nd.connectionTimeOut),
		nats.RetryOnFailedConnect(true),
		nats.ConnectHandler(func(c *nats.Conn) {
			nd.logger.Info().Msgf("Connected to  %s", c.ConnectedUrl())
			go nd.startSubscribe()
		}),
		nats.ClosedHandler(func(c *nats.Conn) {
			nd.logger.Info().Msgf("Connection closed to %s", c.ConnectedUrl())
		}),
		nats.ReconnectWait(nd.connectTimeWait),
		nats.ReconnectHandler(func(c *nats.Conn) {
			nd.logger.Info().Msgf("Reconnected to %s", c.ConnectedUrl())
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
	nc, err := nats.Connect(cfg.NatsServers(), opts...)
	if err != nil {
		return err
	}
	nd.nc = nc
	return nil
}

var sub sync.Mutex

func (nd *NatsDialer) startSubscribe() {
	sub.Lock()

	defer sub.Unlock()

	for _, sub := range nd.subscriptions {
		_ctx, _cancel := context.WithTimeout(nd.ctx, nd.connectionTimeOut)
		nd.logger.Info().Msgf("Start subscribing to %s", sub.String())
		err := sub.Subscribe(_ctx, _cancel, nd.nc)
		if err != nil {
			nd.logger.Error().Err(err).Msgf("Subscription to %s failed", sub.String())
			time.Sleep(nd.connectTimeWait)
			go nd.startSubscribe()
		}
	}
}

func (nd *NatsDialer) startUnSubscribe() {
	sub.Lock()

	defer sub.Unlock()

	for _, sub := range nd.subscriptions {
		nd.logger.Info().Msgf("Start unsubscribing to %s", sub.String())
		err := sub.Unsubscribe()
		if err != nil {
			nd.logger.Error().Err(err).Msgf("Delete subscription to %s failed", sub.String())
		}
	}
}

func (nd *NatsDialer) Disconnect() error {
	nd.startUnSubscribe()
	if nd.nc == nil || !nd.nc.IsConnected() {
		nd.logger.Info().Msg("Not connected to server nothing todo")
		return nil
	}
	// Disconnect and flush pending messages
	if err := nd.nc.Drain(); err != nil {
		nd.logger.Error().Err(err).Msg("Can't Drain")
		return err
	}
	nd.logger.Info().Msg("Disconnected")
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

type NatsSubscription interface {
	String() string
	Subscribe(ctx context.Context, cancel context.CancelFunc, connection *nats.Conn) error
	Unsubscribe() error
}
