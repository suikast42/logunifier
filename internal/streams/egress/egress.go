package egress

import (
	"encoding/json"
	"errors"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"github.com/suikast42/logunifier/cmd/model"
	"github.com/suikast42/logunifier/internal/config"
	"github.com/suikast42/logunifier/internal/streams/converter"
	"sync"
	"time"
)

type MsgContext struct {
	Orig      *nats.Msg
	Converter converter.EcsConverter
}
type Egress struct {
	logger            *zerolog.Logger
	validationChannel <-chan *MsgContext
	ackTimeout        time.Duration
}

var lock = &sync.Mutex{}
var instance *Egress

func Start(validationChannel <-chan *MsgContext) *Egress {
	lock.Lock()
	defer lock.Unlock()
	cfg, _ := config.Instance()
	if instance == nil {
		logger := config.Logger()
		instance = &Egress{
			logger:            &logger,
			validationChannel: validationChannel,
			ackTimeout:        time.Second * time.Duration(cfg.AckTimeoutS()),
		}
		go instance.startReceiving()
	}

	return instance
}

func (eg *Egress) startReceiving() {
	eg.logger.Info().Msgf("Start validation channel")
	for {
		select {
		case receivedCtx, ok := <-eg.validationChannel:
			if !ok {
				eg.logger.Error().Msgf("Nothing received %v %v", receivedCtx, ok)
				return
			}
			err := receivedCtx.Orig.InProgress()
			if err != nil {
				eg.logger.Error().Err(err).Msg("Can't set message InProgress")
				continue
			}
			converted := receivedCtx.Converter.Convert(receivedCtx.Orig)
			err = eg.validate(converted)
			if err != nil {
				// Do not receive again
				err := receivedCtx.Orig.Ack()
				if err != nil {
					eg.logger.Error().Err(err).Msg("Can't ack message")
				}
				continue
			}
			err = receivedCtx.Orig.Ack()
			if err != nil {
				eg.logger.Error().Err(err).Msg("Can't ack message")
			}
			continue

		case <-time.After(eg.ackTimeout):
			eg.logger.Debug().Msgf("Nothing to validate after %v ", eg.ackTimeout)
			continue
		}
	}
}

func (eg *Egress) validate(msg *model.EcsLogEntry) error {
	_, err := json.Marshal(msg)
	if err != nil {
		return errors.New("can't create json from message")
	}
	//eg.logger.Info().Msgf("Received %s", msg.Timestamp.AsTime().String())
	return nil
}
