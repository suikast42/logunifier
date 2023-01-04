package testingress

import (
	"github.com/nats-io/nats.go"
	"github.com/suikast42/logunifier/internal/config"
	"github.com/suikast42/logunifier/internal/streams/ingress"
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
	ts, err := time.Parse(time.RFC3339, string(msg.Data))
	if err != nil {
		ts = time.Now()
	}
	logger.Debug().Msgf("Counter %d. Message %s", r.testCounter, string(msg.Data))

	return &model.EcsLogEntry{
		Id:      model.UUID(),
		Message: string(msg.Data),
		//Timestamp: timestamppb.New(time.Now()),
		Timestamp: timestamppb.New(ts),
		Labels: map[string]string{
			ingress.IndexedLabelIngress:     "vector-testingress",
			ingress.IndexedLabelUsedPattern: "nil",
			ingress.IndexedLabelJob:         "test",
		},
	}

}
