package lokishipper

import (
	"os"
	"os/signal"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/grafana/dskit/backoff"
	"github.com/grafana/dskit/flagext"
	"github.com/grafana/loki/pkg/push"
	"github.com/grafana/loki/v3/pkg/logproto"
	"github.com/nats-io/nats.go"
	pmodel "github.com/prometheus/common/model"
	"github.com/rs/zerolog"
	"github.com/suikast42/logunifier/internal/config"
	"github.com/suikast42/logunifier/internal/streams/connectors"
	"github.com/suikast42/logunifier/pkg/clients"
	"github.com/suikast42/logunifier/pkg/clients/lokiclient"
	lokiapi "github.com/suikast42/logunifier/pkg/clients/lokiclient/api"
	"github.com/suikast42/logunifier/pkg/model"
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
)

type LokiShipper struct {
	//GRPC_GO_LOG_VERBOSITY_LEVEL=99
	//GRPC_GO_LOG_SEVERITY_LEVEL=info
	logger                   zerolog.Logger
	lokiAddresses            []string
	client                   lokiclient.Client
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
	l.LogWithMetadata(toLokiLevel(ecs), labels, ecs.Timestamp.AsTime(), string(marshal), structuredMetadata(ecs), msg)

}

var conSync sync.Mutex

func (app *LokiShipper) LogWithMetadata(level pmodel.LabelValue, lokiLabels pmodel.LabelSet, t time.Time, message string, metadata push.LabelsAdapter, msg *nats.Msg) {
	lvlCtxLabels, ok := app.levelAdapters[level]
	labels := pmodel.LabelSet{}
	if ok {
		labels = lvlCtxLabels.Merge(lokiLabels)
	} else {
		labels = lokiLabels
	}

	entry := &lokiapi.Entry{
		Labels: labels,
		Entry: logproto.Entry{
			Timestamp:          t,
			Line:               message,
			StructuredMetadata: metadata,
		},
	}

	lokiapi.AddFeedbackNotifier(
		entry,
		msg,
		func(e *lokiapi.Entry, c *nats.Msg, status int) {
			err := msg.Ack()
			if err != nil {
				app.Logger().Err(err).Msgf("Can't ack message. ")
			}
		},
		func(e *lokiapi.Entry, c *nats.Msg, status int, lokiSendError error) {
			if lokiSendError != nil && strings.Contains(strings.ToLower(lokiSendError.Error()), "entry too far behind") {
				app.Logger().Error().Err(lokiSendError).Msgf("Event lost. Can't push message to l. Lost message: [%s]", message)
				err := msg.Term()
				if err != nil {
					app.Logger().Error().Err(err).Msg("Can't terminate message")
				}
			} else {
				//l.Logger().Error().Err(pushErr).Msgf("Can't push message to l. %s", marshal)
				err := c.NakWithDelay(app.natsRedeliverInterval)
				if err != nil {
					app.Logger().Error().Err(err).Msg("Can't nack message")
				}
			}
		},
	)
	//app.client.Chan() <- *entry
	app.client.Send(entry)
}

func (l *LokiShipper) Connect() {
	go func(conSync *sync.Mutex) {
		conSync.Lock()

		defer conSync.Unlock()
		// TODO: how to connect to a loki cluster ?
		var url = flagext.URLValue{}
		urlErr := url.Set(l.LokiAddresses()[0] + "/loki/api/v1/push")
		if urlErr != nil {
			panic(urlErr)
		}
		cfg := lokiclient.Config{
			URL:       url,
			BatchWait: 1 * time.Second,
			BatchSize: 32 * 1024 * 1024, // 32MB
			//ExternalLabels: lokiflag.LabelSet{LabelSet: pmodel.LabelSet{"app": "robust-service"}},
			BackoffConfig: backoff.Config{
				MinBackoff: 100 * time.Millisecond,
				MaxBackoff: 100 * time.Millisecond,
				MaxRetries: 1, // High retry count handles longer reboots
			},
			Timeout: 10 * time.Second,
		}

		client, err := clients.NewLokiClient(cfg)
		if err != nil {
			panic(err)
		}
		l.client = client
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
	if ecs.HasErrorType() {
		labelsMap[pmodel.LabelName("error_type")] = pmodel.LabelValue(ecs.Error.Type)
	}
	return labelsMap
}

func toLokiLevel(ecs *model.EcsLogEntry) pmodel.LabelValue {
	return pmodel.LabelValue(ecs.Log.Level.Enum().String())
}
