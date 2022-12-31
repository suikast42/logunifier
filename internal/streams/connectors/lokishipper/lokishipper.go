package lokishipper

// The loki client has an issue with the mod versionin in v2
// see https://github.com/grafana/loki/issues/2826
// For upgrading the loki client version go to GitHub and find aut the commit id
// After that execute go get github.com/grafana/loki@a290549a59fed24a4374e922dafee6c784cdedcc
// This commit  id above resolves a go module github.com/grafana/loki v1.6.2-0.20220914103657-a290549a59fe
// Furthermore, the weavework/common dependency from loki has a incompatibility with the grpc > 1.45
// Thus you must patch it with  go get github.com/weaveworks/common@7c2720a9024d7fdf5f4668321a4fcddf7b461b27
// This os the case with loki client v 2.7.1 with commit id a290549a59fed24a4374e922dafee6c784cdedcc
// Check out if it needed for further releases

// Grafana loki model parsing https://github.com/grafana/loki/issues/114
import (
	"context"
	"encoding/json"
	"github.com/grafana/loki/pkg/logproto"
	"github.com/nats-io/nats.go"
	lokiLabels "github.com/prometheus/prometheus/model/labels"
	"github.com/rs/zerolog"
	"github.com/suikast42/logunifier/pkg/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"sync"
	"time"
)

type LokiShipper struct {
	//GRPC_GO_LOG_VERBOSITY_LEVEL=99
	//GRPC_GO_LOG_SEVERITY_LEVEL=info
	Logger         zerolog.Logger
	LokiAddresses  []string
	grpcConnection *grpc.ClientConn
	client         logproto.PusherClient
	ctx            context.Context
	cancelFnc      context.CancelFunc
	connected      bool
}

func (loki *LokiShipper) Handle(msg *nats.Msg, ecs *model.EcsLogEntry) {
	if !loki.connected {
		//loki.Logger.Debug().Msg("not connected to loki")
		err := msg.NakWithDelay(time.Second * 5)
		if err != nil {
			loki.Logger.Error().Err(err).Msg("Can't nack message")
		}
		return
	}

	labels := toLokiLabels(ecs)
	marshal, marshalError := json.Marshal(ecs)
	if marshalError != nil {
		loki.Logger.Err(marshalError).Msgf("Can't marshal message %+v", ecs)
		err := msg.Term()
		if err != nil {
			loki.Logger.Error().Err(err).Msg("Can't terminate message")
		}
		return
	}
	pushRequest := loki.buildPushRequest(ecs.Timestamp.AsTime(), labels, string(marshal))
	pushResponse, pushErr := loki.client.Push(loki.ctx, pushRequest)
	if pushErr != nil {
		loki.Logger.Err(pushErr).Msg("Can't push message")
		err := msg.NakWithDelay(time.Second * 5)
		if err != nil {
			loki.Logger.Error().Err(err).Msg("Can't nack message")
		}
		return
	}
	err := msg.Ack()
	if err != nil {
		loki.Logger.Err(pushErr).Msgf("Can't ack message", pushResponse)
	}
}

var conSync sync.Mutex

func (loki *LokiShipper) Connect() {
	go func(conSync *sync.Mutex) {
		conSync.Lock()

		defer conSync.Unlock()
		if loki.connected || loki.grpcConnection != nil {
			return
		}
		grpcConnection, err := grpc.Dial("loki.service.consul:9005", grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			var wait = time.Second * 5
			loki.Logger.Error().Err(err).Msgf("Can't create connection to loki %s. Try in %v", loki.LokiAddresses, wait)
			go loki.Connect()
			return
		}
		if loki.grpcConnection == nil {
			ctx, cancel := context.WithCancel(context.Background())
			loki.ctx = ctx
			loki.cancelFnc = cancel
			loki.grpcConnection = grpcConnection
			go loki.watch()
		}
		loki.client = logproto.NewPusherClient(grpcConnection)

	}(&conSync)

}
func (loki *LokiShipper) watch() {
	for {
		select {
		case <-loki.ctx.Done():
			loki.Logger.Info().Msg("Loki watcher stooped")
			return
		case <-time.After(time.Second):
			state := loki.grpcConnection.GetState()
			loki.connected = state == connectivity.Ready
			if !loki.connected {
				loki.Logger.Info().Msgf("Loki connection state: %s", state)
			}
		}
	}
}

var disconnMtx sync.Mutex

func (loki *LokiShipper) DisConnect() {
	disconnMtx.Lock()
	defer disconnMtx.Unlock()
	if loki.grpcConnection != nil {
		loki.cancelFnc()
		err := loki.grpcConnection.Close()
		loki.Logger.Error().Err(err).Msgf("Can't close connection to loki %s", loki.LokiAddresses)
	}
	loki.grpcConnection = nil
}

func (loki *LokiShipper) buildPushRequest(ts time.Time, labels map[string]string, line string) *logproto.PushRequest {
	req := &logproto.PushRequest{}
	_labels := lokiLabels.FromMap(labels)
	req.Streams = append(req.Streams, logproto.Stream{
		Labels: _labels.String(),
		Entries: []logproto.Entry{
			{
				Timestamp: ts,
				Line:      line,
			},
		},
	})

	return req
}

// toLokiLabels extract indexed fields from the log entry
func toLokiLabels(ecs *model.EcsLogEntry) map[string]string {
	labelsMap := make(map[string]string)
	for k, v := range ecs.Labels {
		labelsMap[k] = v
	}
	labelsMap["job"] = ecs.Service.Name
	// The level label is autodetected by grafana log panel
	// Thus we duplicate this
	labelsMap["level"] = model.LogLevelToString(ecs.Log.Level)
	return labelsMap
}
