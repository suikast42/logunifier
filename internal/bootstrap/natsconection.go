package bootstrap

import (
	"context"
	"github.com/nats-io/nats.go"
	"github.com/suikast42/logunifier/internal/config"
	"log"
	"net"
	"time"
)
import _ "github.com/nats-io/nats.go"

// use the custom dialer implementation from the nats tutorial
// https://docs.nats.io/using-nats/developer/tutorials/custom_dialer
type customDialer struct {
	ctx             context.Context
	nc              *nats.Conn
	connectTimeout  time.Duration
	connectTimeWait time.Duration
}

func Connect() error {
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
	}
	opts := []nats.Option{
		nats.SetCustomDialer(cd),
		nats.ReconnectWait(2 * time.Second),
		nats.ReconnectHandler(func(c *nats.Conn) {
			logger.Info().Msgf("Reconnected to %s", c.ConnectedUrl())
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
			logger.Error().Err(err).Msg("")
		}

		// Wait for context to be canceled either by timeout
		// or because of establishing a connection...
		select {
		case <-ctx.Done():
			break WaitForEstablishedConnection
		default:
		}

		if nc == nil || !nc.IsConnected() {
			logger.Warn().Msg("Connection not ready")
			time.Sleep(200 * time.Millisecond)
			continue
		}
		break WaitForEstablishedConnection
	}
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
		logger.Error().Err(err).Msg("")
	}
	logger.Info().Msg("Disconnected")

	//
	//// Normally, the library will return an error when trying to connect and
	//// there is no server running. The RetryOnFailedConnect option will set
	//// the connection in reconnecting state if it failed to connect right away.
	//nc, err := nats.Connect(cfg.NatsServers(),
	//	nats.RetryOnFailedConnect(true),
	//	nats.MaxReconnects(10),
	//	nats.ReconnectWait(time.Second),
	//	nats.ReconnectHandler(func(connection *nats.Conn) {
	//		logger.Info().Msgf("Connecting to to %s", connection.ConnectedUrl())
	//		// Note that this will be invoked for the first asynchronous connect.
	//	}))
	//if err != nil {
	//	// Should not return an error even if it can't connect, but you still
	//	// need to check in case there are some configuration errors.
	//}
	//
	//for !nc.IsConnected() {
	//	time.Sleep(time.Second * 1)
	//}
	//logger.Info().Msgf("Connected to %s", nc.ConnectedUrl())
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
				logger.Info().Msgf("Connected to NATS successfully")
				return conn, nil
			} else {
				time.Sleep(cd.connectTimeWait)
			}
		}
	}
}
