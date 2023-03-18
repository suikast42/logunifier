package process

import (
	"context"
	"encoding/json"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"github.com/suikast42/logunifier/internal/bootstrap"
	"github.com/suikast42/logunifier/internal/config"
	"github.com/suikast42/logunifier/internal/streams/ingress"
	"github.com/suikast42/logunifier/pkg/patterns"
	"os"
	"runtime/debug"
	"sync"
	"time"
)

type LogProcessor struct {
	logger            *zerolog.Logger
	validationChannel <-chan ingress.IngressMsgContext
	ackTimeout        time.Duration
	pushSubject       string
}

var lock = &sync.Mutex{}
var instance *LogProcessor

func Start(processChannel <-chan ingress.IngressMsgContext, pushSubject string) error {
	lock.Lock()
	defer lock.Unlock()
	cfg, _ := config.Instance()
	if instance == nil {
		logger := config.Logger()

		instance = &LogProcessor{
			logger:            &logger,
			validationChannel: processChannel,
			ackTimeout:        time.Second * time.Duration(cfg.AckTimeoutS()),
			pushSubject:       pushSubject,
		}
		go instance.startReceiving()
	}

	return nil
}

func (eg *LogProcessor) startReceiving() {
	defer func() {
		if r := recover(); r != nil {
			// Log fatal do an os.Exit(1)
			logger := config.Logger()
			stack := debug.Stack()
			logger.Fatal().Msgf("Unexpected error: %+v\n%s", r, string(stack))
		}
	}()
	instance, _ := bootstrap.Intance()
	for instance == nil {
		instance, _ = bootstrap.Intance()
		eg.logger.Info().Msgf("Waiting for boostrap is done")
		time.Sleep(time.Second * 1)
	}

	nc := instance.Connection()
	for nc == nil {
		nc = instance.Connection()
		eg.logger.Info().Msgf("Waiting connection to nats established")
		time.Sleep(time.Second * 1)
	}

	egressStream, err := bootstrap.ProducerStream(context.Background(), nc)
	if err != nil {
		eg.logger.Error().Err(err).Msg("Can't create producer for egress stream")
		os.Exit(1)
	}
	eg.logger.Info().Msgf("Start validation channel")
	patternFactory := patterns.Instance()
	for {
		select {
		case receivedCtx, ok := <-eg.validationChannel:
			if !ok {
				instance = nil
				eg.logger.Error().Msgf("Nothing received %v %v", receivedCtx, ok)
				return
			}
			err := receivedCtx.NatsMsg.InProgress()
			if err != nil {
				eg.logger.Error().Err(err).Msg("Can't set message InProgress")
				continue
			}

			ecsLog := patternFactory.Parse(receivedCtx.MetaLog)
			ValidateAndFix(ecsLog, receivedCtx.MetaLog)
			// Delete the debug info if there is no error occured there
			if !ecsLog.HasProcessError() {
				ecsLog.ProcessError = nil
			}
			marshal, err := json.Marshal(ecsLog)

			if err != nil {
				eg.logger.Error().Err(err).Msgf("Can't unmarshal outgoing message: %v", ecsLog)
				err = receivedCtx.NatsMsg.Ack()
				if err != nil {
					eg.logger.Error().Err(err).Msg("Can't ack message")
				}
				continue
			}
			async, err := egressStream.PublishAsync(eg.pushSubject, marshal)
			if err != nil {
				eg.logger.Error().Err(err).Msg("Can't publish message")
				err := receivedCtx.NatsMsg.NakWithDelay(eg.ackTimeout)
				if err != nil {
					eg.logger.Error().Err(err).Msg("Can't nack message. Message lost")
				}
				continue
			}
			go func(ack nats.PubAckFuture, msgctx ingress.IngressMsgContext, ackTimeout time.Duration) {
				select {
				case <-ack.Ok():
					err = msgctx.NatsMsg.Ack()
					if err != nil {
						eg.logger.Error().Err(err).Msg("Can't ack message")
					}
					//eg.logger.Debug().Msg("Msg Acked")
				case err, _ := <-ack.Err():
					eg.logger.Error().Err(err).Msgf("Can't to egress %s. Try to nack with a delay of %v", eg.pushSubject, eg.ackTimeout)
					err = msgctx.NatsMsg.NakWithDelay(eg.ackTimeout)
					if err != nil {
						eg.logger.Error().Err(err).Msg("Can't nack message")
					}
				case <-time.After(ackTimeout + time.Second*1):
					//eg.logger.Error().Msgf("This should not happened. Timeout on send msg after  %v ", ackTimeout+time.Second*1)
					err = msgctx.NatsMsg.NakWithDelay(eg.ackTimeout)
					if err != nil {
						eg.logger.Error().Err(err).Msgf("Can't nack message. Message lost. [%s]", string(msgctx.NatsMsg.Data))
					}
				}

			}(async, receivedCtx, time.Second*2)

		case <-time.After(eg.ackTimeout):
			eg.logger.Debug().Msgf("Nothing to validate after %v ", eg.ackTimeout)
			continue
		}
	}

}
