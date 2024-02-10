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
	"strconv"
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
	// Loki does not support array fields.
	// Merge the tags in tge labels if there's some
	if ecs.HasTags() {
		if ecs.Labels == nil {
			ecs.Labels = map[string]string{}
		}
		for i := 0; i < len(ecs.Tags); i++ {
			val := ecs.Tags[i]
			if len(val) > 0 {
				ecs.Labels["tags_"+strconv.Itoa(i)] = val
			}
		}
	}
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
			case connectivity.TransientFailure, connectivity.Idle, connectivity.Connecting:
				loki.Logger.Info().Msgf("Reconnecting to Loki. State is %s", state)
				go func() {
					// Disconnect will trigger a loki.ctx cancel
					loki.DisConnect()
					loki.Connect()
				}()
			case connectivity.Shutdown:
				loki.Logger.Info().Msgf("Shutting down to . State is %s", state)
				go func() {
					// Disconnect will trigger a loki.ctx cancel
					loki.DisConnect()
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
// ingress , host, org_name ,environment,service_stack, service_name, service_type,  service_type , service_namespace , log_level , pattern
func toLokiLabels(ecs *model.EcsLogEntry) map[string]string {
	labelsMap := make(map[string]string)
	labelsMap["ingress"] = ecs.Log.Ingress
	labelsMap["host"] = ecs.Host.Name
	labelsMap["org_name"] = ecs.Organization.Name
	labelsMap["environment"] = ecs.Environment.Name
	labelsMap["service_stack"] = ecs.Service.Stack
	labelsMap["service_name"] = ecs.Service.Name
	labelsMap["service_type"] = ecs.Service.Type
	labelsMap["service_namespace"] = ecs.Service.Namespace
	labelsMap["log_logger"] = ecs.Log.Logger
	labelsMap["level"] = ecs.Log.Level.String()
	labelsMap["pattern_key"] = ecs.Log.PatternKey

	if ecs.HasProcessError() {
		labelsMap["process_error"] = "true"
	} else {
		labelsMap["process_error"] = "false"
	}

	if ecs.HasValidationError() {
		labelsMap["validation_error"] = "true"
	} else {
		labelsMap["validation_error"] = "false"
	}

	if ecs.HasExceptionStackStrace() {
		labelsMap["error_stack"] = "true"
	} else {
		labelsMap["error_stack"] = "false"
	}

	return labelsMap
}
