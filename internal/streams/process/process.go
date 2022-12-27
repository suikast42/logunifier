package process

import (
	"encoding/json"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"github.com/suikast42/logunifier/internal/config"
	"github.com/suikast42/logunifier/internal/streams/ingress"
	"github.com/suikast42/logunifier/pkg/model"
	"sync"
	"time"
)

type LogProcessor struct {
	logger            *zerolog.Logger
	validationChannel <-chan *ingress.IngressMsgContext
	ackTimeout        time.Duration
	egressDestination string
	egressStream      nats.JetStreamContext
}

var lock = &sync.Mutex{}
var instance *LogProcessor

func Start(processChannel <-chan *ingress.IngressMsgContext, egress string, connection *nats.Conn) error {
	lock.Lock()
	defer lock.Unlock()
	cfg, _ := config.Instance()

	if instance == nil {
		logger := config.Logger()

		// create a Js instance
		f := func(stream nats.JetStream, msg *nats.Msg, _err error) {
			logger.Error().Err(_err).Msgf("PublishAsyncErrHandler: %s. Error in send msg", egress)
		}
		opts := []nats.JSOpt{
			nats.PublishAsyncErrHandler(f),
			nats.PublishAsyncMaxPending(2 * 1024),
		}
		js, err := connection.JetStream(opts...)
		if err != nil {
			logger.Error().Err(err).Msg("Can't create jetstream connection")
			return err
		}

		// create a strem cfg
		streamCfg, err := cfg.EgressStreamCfg()
		if err != nil {
			logger.Error().Err(err).Msg("Can't create egress stream cfg")
			return err
		}
		err = cfg.CreateOrUpdateStream(streamCfg, js)
		if err != nil {
			logger.Error().Err(err).Msg("Can't create egress stream")
			return err
		}

		instance = &LogProcessor{
			logger:            &logger,
			validationChannel: processChannel,
			ackTimeout:        time.Second * time.Duration(cfg.AckTimeoutS()),
			egressDestination: egress,
			egressStream:      js,
		}
		go instance.startReceiving()
	}

	return nil
}

func (eg *LogProcessor) startReceiving() {

	eg.logger.Info().Msgf("Start validation channel")
	for {
		select {
		case receivedCtx, ok := <-eg.validationChannel:
			if !ok {
				instance = nil
				eg.logger.Error().Msgf("Nothing received %v %v", receivedCtx, ok)
				return
			}
			err := receivedCtx.Orig.InProgress()
			if err != nil {
				eg.logger.Error().Err(err).Msg("Can't set message InProgress")
				continue
			}
			converted := receivedCtx.Converter.Convert(receivedCtx.Orig)
			eg.analyze(converted)
			marshal, err := json.Marshal(converted)

			if err != nil {
				eg.logger.Error().Err(err).Msgf("Can't unmarshal outgoing message: %v", converted)
				err = receivedCtx.Orig.Ack()
				if err != nil {
					eg.logger.Error().Err(err).Msg("Can't ack message")
				}
				continue
			}

			async, err := eg.egressStream.PublishAsync(eg.egressDestination, marshal)
			if err != nil {
				eg.logger.Error().Err(err).Msg("Can't publish message")
				err := receivedCtx.Orig.NakWithDelay(eg.ackTimeout)
				if err != nil {
					eg.logger.Error().Err(err).Msg("Can't nack message. Message lost")
				}
				continue
			}
			go func(ack nats.PubAckFuture, msgctx *ingress.IngressMsgContext, ackTimeout time.Duration) {
				select {
				case <-ack.Ok():
					err = msgctx.Orig.Ack()
					if err != nil {
						eg.logger.Error().Err(err).Msg("Can't ack message")
					}
					//eg.logger.Debug().Msg("Msg Acked")
				case err, _ := <-ack.Err():
					eg.logger.Error().Err(err).Msgf("Can't to egress %s. Try to nack with a delay of %v", eg.egressDestination, eg.ackTimeout)
					err = msgctx.Orig.NakWithDelay(eg.ackTimeout)
					if err != nil {
						eg.logger.Error().Err(err).Msg("Can't nack message")
					}
				case <-time.After(ackTimeout + time.Second*1):
					eg.logger.Error().Msgf("This should not happened. Timeout on send msg after  %v ", ackTimeout+time.Second*1)
					err = msgctx.Orig.NakWithDelay(eg.ackTimeout)
					if err != nil {
						eg.logger.Error().Err(err).Msg("Can't nack message")
					}
				}

			}(async, receivedCtx, time.Second*2)

		case <-time.After(eg.ackTimeout):
			eg.logger.Debug().Msgf("Nothing to validate after %v ", eg.ackTimeout)
			continue
		}
	}

}

func (eg *LogProcessor) analyze(msg *model.EcsLogEntry) {
	if msg.HasParseErrors() {
		return
	}
	//eg.logger.Info().Msgf("Received %s", msg.Timestamp.AsTime().String())
}
