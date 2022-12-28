package testingress

import (
	"github.com/nats-io/nats.go"
	"github.com/suikast42/logunifier/internal/config"
	"github.com/suikast42/logunifier/pkg/model"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

type TestEcsConverter struct {
	testCounter int64
}

func (r *TestEcsConverter) Convert(msg *nats.Msg) *model.EcsLogEntry {
	r.testCounter++
	logger := config.Logger()
	logger.Debug().Msgf("Counter %d", r.testCounter)
	return &model.EcsLogEntry{
		Id:        model.UUID(),
		Message:   string(msg.Data),
		Timestamp: timestamppb.New(time.Now()),
	}

}
