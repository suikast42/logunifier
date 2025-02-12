package process

import (
	"context"
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
	logger         *zerolog.Logger
	processChannel <-chan ingress.IngressMsgContext
	ackTimeout     time.Duration
	pushSubject    string
	maxAxPendings  int
	channelName    string
}

var lock = &sync.Mutex{}

//var instance *LogProcessor

func Start(processChannel <-chan ingress.IngressMsgContext, channelName string, pushSubject string, maxAxPendings int) error {
	lock.Lock()
	defer lock.Unlock()
	cfg, _ := config.Instance()
	//if instance == nil {
	logger := config.Logger()

	instance := &LogProcessor{
		logger:         &logger,
		processChannel: processChannel,
		ackTimeout:     time.Second * time.Duration(cfg.AckTimeoutS()),
		pushSubject:    pushSubject,
		maxAxPendings:  maxAxPendings,
		channelName:    channelName,
	}
	go instance.startReceiving()

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
		eg.logger.Info().Msg("Waiting for boostrap is done")
		time.Sleep(time.Second * 1)
	}

	nc := instance.ProducerConnection()
	for nc == nil {
		nc = instance.ProducerConnection()
		eg.logger.Info().Msg("Waiting connection to nats established")
		time.Sleep(time.Second * 1)
	}

	egressStream, err := bootstrap.ProducerStream(context.Background(), nc, eg.maxAxPendings)
	if err != nil {
		eg.logger.Error().Err(err).Msg("Can't create producer for egress stream")
		os.Exit(1)
	}
	eg.logger.Info().Msgf("Start receiving channel for %s", eg.channelName)
	patternFactory := patterns.Instance()
	for {
		select {
		case receivedCtx, ok := <-eg.processChannel:
			if !ok {
				instance = nil
				eg.logger.Error().Msgf("Processor %s Nothing received %v %v", eg.channelName, receivedCtx, ok)
				return
			}
			ecsLog := patternFactory.Parse(receivedCtx.MetaLog)
			ValidateAndFix(ecsLog, receivedCtx.NatsMsg)

			marshal, err := ecsLog.ToJson()

			if err != nil {
				eg.logger.Error().Err(err).Msgf("Can't unmarshal outgoing message: %v", ecsLog)
				err = receivedCtx.NatsMsg.Ack()
				if err != nil {
					eg.logger.Error().Err(err).Msg("Can't ack message")
				}
				continue
			}
			ack, sendErr := egressStream.PublishAsync(eg.pushSubject, marshal)
			if sendErr != nil {
				eg.logger.Error().Err(sendErr).Msg("Can't publish message")
				ackErr := receivedCtx.NatsMsg.NakWithDelay(eg.ackTimeout)
				if ackErr != nil {
					eg.logger.Error().Err(ackErr).Msg("Can't nack message. Message lost")
				}
				continue
			}
			select {
			case _ack := <-ack.Ok():
				err = receivedCtx.NatsMsg.Ack()
				if err != nil {
					eg.logger.Error().Err(err).Msg("Can't ack message")
				}
				if _ack.Duplicate {
					eg.logger.Debug().Msg("Duplicate message ")
				}

			case err, _ := <-ack.Err():
				eg.logger.Error().Err(err).Msgf("Can't to egress %s. Try to nack with a delay of %v", eg.pushSubject, eg.ackTimeout)
				err = receivedCtx.NatsMsg.NakWithDelay(eg.ackTimeout)
				if err != nil {
					eg.logger.Error().Err(err).Msg("Can't nack message")
				}
			case <-time.After(eg.ackTimeout + 1*time.Second):
				eg.logger.Error().Msgf("This should not happened. Timeout on send msg after  %v ", eg.ackTimeout+time.Second*1)
				err = receivedCtx.NatsMsg.NakWithDelay(eg.ackTimeout)
				if err != nil {
					eg.logger.Error().Err(err).Msgf("Can't nack message. Message lost. [%s]", string(receivedCtx.NatsMsg.Data))
				}
			}
		case <-time.After(eg.ackTimeout):
			eg.logger.Warn().Msgf("Processor %s Nothing received after %v", eg.channelName, eg.ackTimeout)
			continue
		}
	}

}
