package lokishipper

import (
	"github.com/grafana/loki/pkg/logproto"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"github.com/suikast42/logunifier/internal/config"
	"github.com/suikast42/logunifier/internal/streams/connectors"
	"github.com/suikast42/logunifier/pkg/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

// The loki client has an issue with the mod version in v2
// see https://github.com/grafana/loki/issues/2826
// For upgrading the loki client version go to GitHub and find out the commit id
// After that execute
// 1. go get github.com/grafana/loki@2535f9bedeae5f27abdbfaf0cc1a8e9f91b6c96d
// 2. go get github.com/grafana/loki/pkg/push@2535f9bedeae5f27abdbfaf0cc1a8e9f91b6c96d
// This commit  id above resolves a go module github.com/grafana/loki v1.6.2-0.20231211180320-2535f9bedeae
// For loki v 2.9.3
// Furthermore, the weavework/common dependency from loki has a incompatibility with the grpc > 1.45
// Thus you must patch it with go get github.com/weaveworks/common@e2613bee6b73c78d2038e248e52fcc824dfe02d0
// Grafana loki model parsing https://github.com/grafana/loki/issues/114
import (
	"context"
	"github.com/pkg/errors"
	lokiLabels "github.com/prometheus/prometheus/model/labels"
)

type LokiShipper struct {
	//GRPC_GO_LOG_VERBOSITY_LEVEL=99
	//GRPC_GO_LOG_SEVERITY_LEVEL=info
	Logger                   zerolog.Logger
	LokiAddresses            []string
	grpcConnection           *grpc.ClientConn
	client                   logproto.PusherClient
	ctx                      context.Context
	cancelFnc                context.CancelFunc
	connected                bool
	lokiReconnectionInterval time.Duration
	natsRedeliverInterval    time.Duration
	AckTimeout               time.Duration
}

func (loki *LokiShipper) Health(ctx context.Context) error {
	if !loki.connected {
		return errors.New("not connected yet")
	}
	return nil
}

var lock = &sync.Mutex{}

func (loki *LokiShipper) StartReceive(processChannel <-chan connectors.EgressMsgContext) {
	lock.Lock()
	defer lock.Unlock()

	defer func() {
		if r := recover(); r != nil {
			// Log fatal do an os.Exit(1)
			logger := config.Logger()
			stack := debug.Stack()
			logger.Fatal().Msgf("Unexpected error: %+v\n%s", r, string(stack))
		}
	}()

	go func() {
		logger := config.Logger()
		for {
			select {
			case receivedCtx, ok := <-processChannel:
				if !ok {
					logger.Error().Msgf("Lokishipper. Nothing received %v %v", receivedCtx, ok)
					return
				}
				err := receivedCtx.NatsMsg.InProgress()
				if err != nil {
					logger.Error().Err(err).Msg("Lokishipper. Can't set message InProgress")
					continue
				}
				go loki.Handle(receivedCtx.NatsMsg, receivedCtx.Ecs)
			case <-time.After(loki.AckTimeout):
				logger.Warn().Msgf("Lokishipper. Nothing received after %v ", loki.AckTimeout)
				continue

			}
		}
	}()

}

func (loki *LokiShipper) Handle(msg *nats.Msg, ecs *model.EcsLogEntry) {
	if !loki.connected {
		//loki.Logger.Debug().Msg("not connected to loki")
		err := msg.NakWithDelay(loki.natsRedeliverInterval)
		if err != nil {
			loki.Logger.Error().Err(err).Msg("Can't nack message")
		}
		return
	}

	labels := toLokiLabels(ecs)
	marshal, marshalError := ecs.ToJson()
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
		//"entry too far behind, the oldest acceptable timestamp is: " + m.cutoff.Format(time.RFC3339)
		//if chunkenc.IsErrTooFarBehind(pushErr) {
		if strings.Contains(strings.ToLower(pushErr.Error()), "entry too far behind") {
			loki.Logger.Error().Err(pushErr).Msgf("Event lost. Can't push message to loki. Lost message: [%s]", marshal)
			err := msg.Term()
			if err != nil {
				loki.Logger.Error().Err(err).Msg("Can't terminate message")
			}
		} else {
			//loki.Logger.Error().Err(pushErr).Msgf("Can't push message to loki. %s", marshal)
			err := msg.NakWithDelay(loki.natsRedeliverInterval)
			if err != nil {
				loki.Logger.Error().Err(err).Msg("Can't nack message")
			}
		}

		return
	}
	err := msg.Ack()
	if err != nil {
		loki.Logger.Err(pushErr).Msgf("Can't ack message. Push response %s", pushResponse)
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
		cfg, _ := config.Instance()
		loki.lokiReconnectionInterval = time.Second * 5
		loki.natsRedeliverInterval = loki.lokiReconnectionInterval + time.Second*5
		grpcConnection, err := grpc.Dial(cfg.LokiServers()[0], grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			loki.Logger.Error().Err(err).Msgf("Can't create connection to loki %s. Try in %v", loki.LokiAddresses, loki.lokiReconnectionInterval)
			time.Sleep(loki.lokiReconnectionInterval)
			go loki.Connect()
			return
		}

		ctx, cancel := context.WithCancel(context.Background())
		loki.ctx = ctx
		loki.cancelFnc = cancel
		loki.grpcConnection = grpcConnection
		loki.client = logproto.NewPusherClient(grpcConnection)
		go loki.watch()

	}(&conSync)

}
func (loki *LokiShipper) watch() {
	loki.Logger.Info().Msg("Loki watcher is started")
	for {
		select {
		case <-loki.ctx.Done():
			loki.Logger.Info().Msg("Loki watcher is stopped")
			return
		case <-time.After(loki.lokiReconnectionInterval):
			state := loki.grpcConnection.GetState()
			//loki.Logger.Debug().Msgf("Loki Watcher is running. State is %s", state)
			switch state {
			case connectivity.Ready:
				if !loki.connected {
					loki.Logger.Info().Msgf("Connected to Loki. State is %s", state)
				}
				loki.connected = true
			case connectivity.TransientFailure, connectivity.Idle:
				loki.Logger.Info().Msgf("Reconnecting to Loki. State is %s", state)
				go func() {
					// Disconnect will trigger a loki.ctx cancel
					loki.DisConnect()
					loki.Connect()
				}()
			}

		}
	}
}

func (loki *LokiShipper) DisConnect() {
	conSync.Lock()
	defer conSync.Unlock()
	if loki.grpcConnection != nil {
		loki.connected = false
		loki.cancelFnc()
		err := loki.grpcConnection.Close()
		if err != nil {
			loki.Logger.Error().Err(err).Msgf("Can't close connection to loki %s", loki.LokiAddresses)
		}
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

// toLokiLabels extract loki index labels from ecs log labels
// ingress , host, job ,job_type, task, task_group,  job_type , namespace ,  stack , level , used_grok
func toLokiLabels(ecs *model.EcsLogEntry) map[string]string {
	labelsMap := make(map[string]string)

	// Ingress index for loki
	extractLabel(labelsMap, ecs, string(model.StaticLabelIngress))
	extractLabel(labelsMap, ecs, string(model.StaticLabelHost))
	extractLabel(labelsMap, ecs, string(model.StaticLabelJob))
	extractLabel(labelsMap, ecs, string(model.StaticLabelJobType))
	extractLabel(labelsMap, ecs, string(model.StaticLabelTask))
	extractLabel(labelsMap, ecs, string(model.StaticLabelTaskGroup))
	extractLabel(labelsMap, ecs, string(model.StaticLabelNameSpace))
	extractLabel(labelsMap, ecs, string(model.StaticLabelStack))
	extractLabel(labelsMap, ecs, string(model.DynamicLabelUsedGrok))

	// The level label is autodetected by grafana log panel
	// Thus we duplicate this
	extractLabelWithDefault(labelsMap, ecs, string(model.DynamicLabelLevel), model.LogLevelToString(ecs.Log.Level))

	//extractLabelIgnoreWhen(labelsMap, ecs, string(model.ContainerName))
	//extractLabelIgnoreWhen(labelsMap, ecs, string(model.ContainerImageName))
	//extractLabelIgnoreWhen(labelsMap, ecs, string(model.ContainerImageRevision))

	if ecs.HasProcessError() {
		labelsMap[string(model.StaticLabelProcessError)] = "true"
	} else {
		labelsMap[string(model.StaticLabelProcessError)] = "false"
	}
	return labelsMap
}

func extractLabel(_map map[string]string, ecs *model.EcsLogEntry, key string) {
	extractLabelWithDefault(_map, ecs, key, "NotDefined")
}

func extractLabelIgnoreWhen(_map map[string]string, ecs *model.EcsLogEntry, key string) {
	extractLabelWithDefault(_map, ecs, key, "")
}

func extractLabelWithDefault(_map map[string]string, ecs *model.EcsLogEntry, key string, _default string) {
	if val, ok := ecs.Labels[key]; ok {
		_map[key] = val
	} else {
		if len(_default) > 0 {
			_map[key] = _default
		}
	}
}
