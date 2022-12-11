package bootstrap

import (
	"context"
	"github.com/nats-io/nats.go"
	"github.com/suikast42/logunifier/internal/config"
	"log"
	"net"
	"sync"
	"time"
)
import _ "github.com/nats-io/nats.go"

// use the custom dialer implementation from the nats tutorial
// https://docs.nats.io/using-nats/developer/tutorials/custom_dialer
type customDialer struct {
	ctx             context.Context
	nc              *nats.Conn
	subscriptions   *[]NatsSubscription
	connectTimeout  time.Duration
	connectTimeWait time.Duration
}

func Connect(subscriptions *[]NatsSubscription) error {
	cfg, err := config.Instance()
	logger := config.Logger()
	if err != nil {
		return nil
	}

	// Parent context cancels connecting/reconnecting altogether.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var nc *nats.Conn
	cd := &customDialer{
		ctx:             ctx,
		connectTimeout:  10 * time.Second,
		connectTimeWait: 3 * time.Second,
		subscriptions:   subscriptions,
	}
	opts := []nats.Option{
		nats.SetCustomDialer(cd),
		nats.ReconnectWait(2 * time.Second),
		nats.ReconnectHandler(func(c *nats.Conn) {
			cd.nc = nc
			logger.Info().Msgf("Reconnected to %s", c.ConnectedUrl())
			go cd.startSubscribe()
		}),
		nats.DisconnectErrHandler(func(c *nats.Conn, disconnectionError error) {
			if disconnectionError != nil {
				logger.Error().Err(disconnectionError).Msg("Disconnection with error")
			} else {
				logger.Info().Msg("Disconnected from NATS")
			}
		}),
		nats.ClosedHandler(func(c *nats.Conn) {
			logger.Info().Msg("NATS connection is closed.")
		}),
		//This will kill the client if the connection is lost
		//We will keep the connection
		//nats.NoReconnect(),
	}
	go func() {
		nc, err = nats.Connect(cfg.NatsServers(), opts...)
	}()

WaitForEstablishedConnection:
	for {
		if err != nil {
			logger.Error().Err(err).Msg("Connection error")
			err = nil
		}

		// Wait for context to be canceled either by timeout
		// or because of establishing a connection...
		select {
		case <-ctx.Done():
			break WaitForEstablishedConnection
		default:
		}

		if nc == nil || !nc.IsConnected() {
			if nc == nil {
				logger.Warn().Msg("Connection not ready ")
			} else {
				logger.Warn().Msgf("Connection not ready %v ", nc)
			}
			time.Sleep(cd.connectTimeWait)
			continue
		}

		break WaitForEstablishedConnection
	}
	cd.nc = nc
	go cd.startSubscribe()
	if ctx.Err() != nil {
		log.Fatal(ctx.Err())
	}

	for {
		if nc.IsClosed() {
			break
		}
		if err := nc.Publish("hello", []byte("world")); err != nil {
			logger.Error().Err(err).Msg("")
			time.Sleep(1 * time.Second)
			continue
		}
		logger.Debug().Msg("Published message")
		time.Sleep(1 * time.Second)
	}

	// Disconnect and flush pending messages
	if err := nc.Drain(); err != nil {
		logger.Error().Err(err).Msg("Can't Drain")
	}
	logger.Info().Msg("Disconnected")

	return nil
}

func (cd *customDialer) Dial(network, address string) (net.Conn, error) {
	ctx, cancel := context.WithTimeout(cd.ctx, cd.connectTimeout)
	logger := config.Logger()
	defer cancel()

	for {
		logger.Info().Msgf("Attempting to connect to %s", address)
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		select {
		case <-cd.ctx.Done():
			return nil, cd.ctx.Err()
		default:
			d := &net.Dialer{}
			if conn, err := d.DialContext(ctx, network, address); err == nil {
				logger.Info().Msgf("Connected to NATS %s successfully", address)

				return conn, nil
			} else {
				time.Sleep(cd.connectTimeWait)
			}
		}
	}
}

var sub sync.Mutex

func (cd customDialer) startSubscribe() {
	sub.Lock()
	logger := config.Logger()
	defer sub.Unlock()

	for _, sub := range *cd.subscriptions {
		_ctx, _cancel := context.WithTimeout(cd.ctx, cd.connectTimeout)
		err := sub.Subscribe(_ctx, _cancel, cd.nc)
		if err != nil {
			logger.Error().Err(err).Msgf("Subscription to %s failed", sub.String())
			time.Sleep(cd.connectTimeWait)
			go cd.startSubscribe()
		}
	}
}

type NatsSubscription interface {
	String() string
	Subscribe(ctx context.Context, cancel context.CancelFunc, connection *nats.Conn) error
}
