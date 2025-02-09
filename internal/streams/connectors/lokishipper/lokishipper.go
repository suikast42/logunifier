package lokishipper

import (
	"github.com/grafana/loki/pkg/push"
	"github.com/grafana/loki/v3/pkg/logproto"
	"github.com/nats-io/nats.go"
	pmodel "github.com/prometheus/common/model"
	"github.com/rs/zerolog"
	"github.com/suikast42/logunifier/internal/config"
	"github.com/suikast42/logunifier/internal/streams/connectors"
	"github.com/suikast42/logunifier/pkg/model"
	"os"
	"os/signal"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"
)

// !!! The module issue is fixed with the V3 version of loki !!!
//The loki client has an issue with the mod version in v2
// see https://github.com/grafana/loki/issues/2826
// For upgrading the loki client version go to GitHub and find out the commit id
// After that execute
// 1. go get github.com/grafana/loki@318652035059fdaa40405f263fc9e37b4d38b157
// 2. go get github.com/grafana/loki/pkg/push@318652035059fdaa40405f263fc9e37b4d38b157
// This commit  id above resolves a go module github.com/grafana/loki v1.6.2-0.20231211180320-2535f9bedeae
// For loki v 2.9.6
// Furthermore, the weavework/common dependency from loki has a incompatibility with the grpc > 1.45
// Thus you must patch it with go get github.com/weaveworks/common@e2613bee6b73c78d2038e248e52fcc824dfe02d0

// Grafana loki model parsing https://github.com/grafana/loki/issues/114
import (
	"context"
	lokiLabels "github.com/prometheus/prometheus/model/labels"

	"github.com/grafana/loki-client-go/loki"
	_ "github.com/grafana/loki/pkg/push"
	_ "github.com/prometheus/common/model"
)

type LokiShipper struct {
	//GRPC_GO_LOG_VERBOSITY_LEVEL=99
	//GRPC_GO_LOG_SEVERITY_LEVEL=info
	logger                   zerolog.Logger
	lokiAddresses            []string
	client                   *loki.Client
	ctx                      context.Context
	cancelFnc                context.CancelFunc
	lokiReconnectionInterval time.Duration
	natsRedeliverInterval    time.Duration
	ackTimeout               time.Duration
	levelAdapters            map[pmodel.LabelValue]pmodel.LabelSet
}

func NewLokiShipper(cfg *config.Config) *LokiShipper {
	adapters := make(map[pmodel.LabelValue]pmodel.LabelSet)
	_labels := make(pmodel.LabelSet)
	levels := model.LogLevels()
	for _, v := range levels {
		adapters[pmodel.LabelValue(v.String())] = _labels.Merge(pmodel.LabelSet{"level": pmodel.LabelValue(v.String())})
	}
	return &LokiShipper{
		logger:        config.Logger(),
		lokiAddresses: cfg.LokiServers(),
		ackTimeout:    time.Second * time.Duration(cfg.AckTimeoutS()),
		levelAdapters: adapters,
	}
}
func (l *LokiShipper) Health(ctx context.Context) error {
	return nil
}

var lock = &sync.Mutex{}

func (l *LokiShipper) LokiAddresses() []string {
	return l.lokiAddresses
}
func (l *LokiShipper) StartReceive(processChannel <-chan connectors.EgressMsgContext) {
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
				//err := receivedCtx.NatsMsg.InProgress()
				//if err != nil {
				//	logger.Error().Err(err).Msg("Lokishipper. Can't set message InProgress")
				//	continue
				//}
				go l.Handle(receivedCtx.NatsMsg, receivedCtx.Ecs)
			case <-time.After(l.ackTimeout):
				logger.Warn().Msgf("Lokishipper. Nothing received after %v ", l.ackTimeout)
				continue

			}
		}
	}()

}

func (l *LokiShipper) Logger() *zerolog.Logger {
	return &l.logger
}
func (l *LokiShipper) Handle(msg *nats.Msg, ecs *model.EcsLogEntry) {

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
		l.Logger().Err(marshalError).Msgf("Can't marshal message %+v", ecs)
		err := msg.Term()
		if err != nil {
			l.Logger().Error().Err(err).Msg("Can't terminate message")
		}
		return
	}
	lokiSendError := l.LogWithMetadata(toLokiLevel(ecs), labels, ecs.Timestamp.AsTime(), string(marshal), structuredMetadata(ecs))
	if lokiSendError != nil {
		if strings.Contains(strings.ToLower(lokiSendError.Error()), "entry too far behind") {
			l.Logger().Error().Err(lokiSendError).Msgf("Event lost. Can't push message to l. Lost message: [%s]", marshal)
			err := msg.Term()
			if err != nil {
				l.Logger().Error().Err(err).Msg("Can't terminate message")
			}
		} else {
			//l.Logger().Error().Err(pushErr).Msgf("Can't push message to l. %s", marshal)
			err := msg.NakWithDelay(l.natsRedeliverInterval)
			if err != nil {
				l.Logger().Error().Err(err).Msg("Can't nack message")
			}
		}
		return
	}

	err := msg.Ack()
	if err != nil {
		l.Logger().Err(lokiSendError).Msgf("Can't ack message. ")
	}
}

var conSync sync.Mutex

func (app *LokiShipper) LogWithMetadata(level pmodel.LabelValue, lokiLabels pmodel.LabelSet, t time.Time, message string, metadata push.LabelsAdapter) error {
	lvlCtxLabels, ok := app.levelAdapters[level]
	labels := pmodel.LabelSet{}
	if ok {
		labels = lvlCtxLabels.Merge(lokiLabels)
	} else {
		labels = lokiLabels
	}
	err := app.client.HandleWithMetadata(labels, t, message, metadata)
	if err != nil {
		return err
	}
	return nil
}

func (l *LokiShipper) Connect() {
	go func(conSync *sync.Mutex) {
		conSync.Lock()

		defer conSync.Unlock()
		// TODO: how to connect to a loki cluster ?
		cfg, err := loki.NewDefaultConfig(l.LokiAddresses()[0] + "/loki/api/v1/push")
		if err != nil {
			panic(err)
		}
		cfg.BackoffConfig.MaxRetries = 1
		cfg.BackoffConfig.MinBackoff = 100 * time.Millisecond
		cfg.BackoffConfig.MaxBackoff = 100 * time.Millisecond
		client, err := loki.New(cfg)
		l.client = client
		if err != nil {
			panic(err)
		}
		//defer client.Stop()
		//
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		//defer stop()

		l.cancelFnc = stop
		l.ctx = ctx

	}(&conSync)

}

func (l *LokiShipper) DisConnect() {
	l.cancelFnc()
	l.client.Stop()
	//conSync.Lock()
	//defer conSync.Unlock()
	//l.client.Stop()
}

func (l *LokiShipper) buildPushRequest(ts time.Time, labels map[string]string, line string) *logproto.PushRequest {
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

func structuredMetadata(ecs *model.EcsLogEntry) push.LabelsAdapter {
	adapters := make([]push.LabelAdapter, 0)
	if ecs.IsTraceIdSet() {
		labelAdapter := push.LabelAdapter{Name: "traceID", Value: ecs.Trace.Trace.Id}
		adapters = append(adapters, labelAdapter)
	}
	if ecs.IsSpanIdSet() {
		labelAdapter := push.LabelAdapter{Name: "spanID", Value: ecs.Trace.Span.Id}
		adapters = append(adapters, labelAdapter)
	}
	if ecs.IsUserSet() {
		labelAdapter := push.LabelAdapter{Name: "user", Value: ecs.User.Name}
		adapters = append(adapters, labelAdapter)
	}
	return adapters
}

// toLokiLabels extract loki index labels from ecs log labels
// ingress , host, org_name ,environment,service_stack, service_name, service_type,  service_type , service_namespace , log_level , pattern
func toLokiLabels(ecs *model.EcsLogEntry) pmodel.LabelSet {
	labelsMap := make(pmodel.LabelSet)
	labelsMap[pmodel.LabelName("ingress")] = pmodel.LabelValue(ecs.Log.Ingress)
	labelsMap[pmodel.LabelName("host")] = pmodel.LabelValue(ecs.Host.Name)
	labelsMap[pmodel.LabelName("org_name")] = pmodel.LabelValue(ecs.Organization.Name)
	labelsMap[pmodel.LabelName("environment")] = pmodel.LabelValue(ecs.Environment.Name)
	labelsMap[pmodel.LabelName("service_stack")] = pmodel.LabelValue(ecs.Service.Stack)
	labelsMap[pmodel.LabelName("service_name")] = pmodel.LabelValue(ecs.Service.Name)
	labelsMap[pmodel.LabelName("service_type")] = pmodel.LabelValue(ecs.Service.Type)
	labelsMap[pmodel.LabelName("service_namespace")] = pmodel.LabelValue(ecs.Service.Namespace)
	labelsMap[pmodel.LabelName("log_logger")] = pmodel.LabelValue(ecs.Log.Logger)
	labelsMap[pmodel.LabelName("level")] = pmodel.LabelValue(ecs.Log.Level.String())
	labelsMap[pmodel.LabelName("pattern_key")] = pmodel.LabelValue(ecs.Log.PatternKey)
	labelsMap[pmodel.LabelName("process_error")] = pmodel.LabelValue(strconv.FormatBool(ecs.HasProcessError()))
	labelsMap[pmodel.LabelName("validation_error")] = pmodel.LabelValue(strconv.FormatBool(ecs.HasValidationError()))
	labelsMap[pmodel.LabelName("error_stack")] = pmodel.LabelValue(strconv.FormatBool(ecs.HasExceptionStackStrace()))

	return labelsMap
}

func toLokiLevel(ecs *model.EcsLogEntry) pmodel.LabelValue {
	return pmodel.LabelValue(ecs.Log.Level.Enum().String())
}
