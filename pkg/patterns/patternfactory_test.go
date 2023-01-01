package patterns

import (
	"github.com/rs/zerolog"
	"github.com/suikast42/logunifier/internal/config"
	"github.com/suikast42/logunifier/pkg/model"
	"os"
	"reflect"
	"testing"
	"time"
)

var (
	patternfactory *PatternFactory
	logger         zerolog.Logger
)

func init() {
	logger = config.Logger()
	_, err := Initialize()
	if err != nil {
		logger.Error().Err(err).Stack().Msg("Can't initialize pattern factory")
		os.Exit(1)
	}
	patternfactory = Instance()
}
func TestParseTS_LEVEL(t *testing.T) {
	{
		var log = "2022-12-08T12:21:02.594Z [ERROR] nomad.autopilot: 😜 Failed\nto reconcile current state with the desired state\nthird line mf\n1\n3"
		ts, _ := time.Parse(time.RFC3339, "2022-12-08T12:21:02.594Z")
		expected := ParseResult{
			LogLevel:    "ERROR",
			TimeStamp:   ts,
			UsedPattern: string(TS_LEVEL),
		}
		parsed, err := patternfactory.Parse(TS_LEVEL, log)

		if err != nil {
			t.Error(err)
		}
		equal := reflect.DeepEqual(expected, parsed)
		if !equal {
			t.Errorf("Expected %+v but got %+v", expected, parsed)
		}
	}
	{
		var log = "2022-12-08T12:21:02.594Z ERROR nomad.autopilot: 😜 Failed\nto reconcile current state with the desired state\nthird line mf\n1\n3"
		ts, _ := time.Parse(time.RFC3339, "2022-12-08T12:21:02.594Z")
		expected := ParseResult{
			LogLevel:    "ERROR",
			TimeStamp:   ts,
			UsedPattern: string(TS_LEVEL),
		}
		parsed, err := patternfactory.Parse(TS_LEVEL, log)

		if err != nil {
			t.Error(err)
		}
		equal := reflect.DeepEqual(expected, parsed)
		if !equal {
			t.Errorf("Expected %+v but got %+v", expected, parsed)
		}
	}
	{
		var log = "2022-12-08T12:21:02.594Z \"ERROR\" nomad.autopilot: 😜 Failed\nto reconcile current state with the desired state\nthird line mf\n1\n3"
		ts, _ := time.Parse(time.RFC3339, "2022-12-08T12:21:02.594Z")
		expected := ParseResult{
			LogLevel:    "ERROR",
			TimeStamp:   ts,
			UsedPattern: string(TS_LEVEL),
		}
		parsed, err := patternfactory.Parse(TS_LEVEL, log)

		if err != nil {
			t.Error(err)
		}
		equal := reflect.DeepEqual(expected, parsed)
		if !equal {
			t.Errorf("Expected %+v but got %+v", expected, parsed)
		}
	}

	{
		var log = "2023-01-01 21:22:59,772 ERROR [org.keycloak.services] (main) KC-SERVICES0010: Failed to add user 'admin' to realm 'master': user with username exists"
		ts, _ := time.Parse(time.RFC3339, "2023-01-01 21:22:59,772")
		expected := ParseResult{
			LogLevel:    "ERROR",
			TimeStamp:   ts,
			UsedPattern: string(TS_LEVEL),
		}
		parsed, err := patternfactory.Parse(TS_LEVEL, log)

		if err != nil {
			t.Error(err)
		}
		equal := reflect.DeepEqual(expected, parsed)
		if !equal {
			t.Errorf("Expected %+v but got %+v", expected, parsed)
		}
	}

	//logger.Info().Msgf("%s", parsed)
}

func TestParseLOGFMT_TS_LEVEL_MSG(t *testing.T) {
	{
		var log = "time=\"2022-12-31T15:55:54.762121247Z\" level=warning msg=\"got error while decoding json\" error=\"unexpected EOF\" retries=1"
		ts, _ := time.Parse(time.RFC3339, "2022-12-31T15:55:54.762121247Z")

		expected := ParseResult{
			TimeStamp:   ts,
			LogLevel:    "warning",
			UsedPattern: string(LOGFMT_TS_LEVEL),
		}
		parsed, err := patternfactory.Parse(LOGFMT_TS_LEVEL, log)

		if err != nil {
			t.Error(err)
		}
		equal := reflect.DeepEqual(expected, parsed)
		if !equal {
			t.Errorf("Expected %+v but got %+v", expected, parsed)
		}
	}
	{
		var log = "ts=\"2022-12-31T15:55:54.762121247Z\" level=warning msg=\"got error while decoding json\" error=\"unexpected EOF\" retries=1"
		ts, _ := time.Parse(time.RFC3339, "2022-12-31T15:55:54.762121247Z")

		expected := ParseResult{
			TimeStamp:   ts,
			LogLevel:    "warning",
			UsedPattern: string(LOGFMT_TS_LEVEL),
		}
		parsed, err := patternfactory.Parse(LOGFMT_TS_LEVEL, log)

		if err != nil {
			t.Error(err)
		}
		equal := reflect.DeepEqual(expected, parsed)
		if !equal {
			t.Errorf("Expected %+v but got %+v", expected, parsed)
		}
	}
	{
		var log = "ts=\"2022-12-31T15:55:54.762121247Z\" level=warning message=\"got error while decoding json\" error=\"unexpected EOF\" retries=1"
		ts, _ := time.Parse(time.RFC3339, "2022-12-31T15:55:54.762121247Z")

		expected := ParseResult{
			TimeStamp:   ts,
			LogLevel:    "warning",
			UsedPattern: string(LOGFMT_TS_LEVEL),
		}
		parsed, err := patternfactory.Parse(LOGFMT_TS_LEVEL, log)

		if err != nil {
			t.Error(err)
		}
		equal := reflect.DeepEqual(expected, parsed)
		if !equal {
			t.Errorf("Expected %+v but got %+v", expected, parsed)
		}
	}
	{
		var log = "ts=2023-01-01T21:44:41.706702241Z caller=spanlogger.go:80 user=anonymous level=debug event=\"complete commit\" commitDuration=82.097µs"
		ts, _ := time.Parse(time.RFC3339, "2023-01-01T21:44:41.706702241Z")

		expected := ParseResult{
			TimeStamp:   ts,
			LogLevel:    "debug",
			UsedPattern: string(LOGFMT_TS_LEVEL),
		}
		parsed, err := patternfactory.Parse(LOGFMT_TS_LEVEL, log)

		if err != nil {
			t.Error(err)
		}
		equal := reflect.DeepEqual(expected, parsed)
		if !equal {
			t.Errorf("Expected %+v but got %+v", expected, parsed)
		}
	}

	{
		var log = "ts=2023-01-01T22:14:36.47634233Z caller=spanlogger.go:80 user=anonymous level=debug event=\"complete commit\" commitDuration=121.065µs"
		ts, _ := time.Parse(time.RFC3339, "2023-01-01T22:14:36.47634233Z")
		expected := ParseResult{
			LogLevel:    "debug",
			TimeStamp:   ts,
			UsedPattern: string(LOGFMT_TS_LEVEL),
		}
		parsed, err := patternfactory.Parse(LOGFMT_TS_LEVEL, log)

		if err != nil {
			t.Error(err)
		}
		equal := reflect.DeepEqual(expected, parsed)
		if !equal {
			t.Errorf("Expected %+v but got %+v", expected, parsed)
		}
	}

	{
		var log = "logger=context userId=0 orgId=0 uname= t=2023-01-01T22:16:44.602022589Z level=info msg=\"Request Completed\" method=GET path=/login/generic_oauth status=302 remote_addr=10.21.21.42 time_ms=0 duration=436.124µs size=304 referer=http://10.21.21.42:29580/login handler=/login/:name/api/v1/push (200) 2.543513ms\""
		ts, _ := time.Parse(time.RFC3339, "2023-01-01T22:16:44.602022589Z")

		expected := ParseResult{
			TimeStamp:   ts,
			LogLevel:    "info",
			UsedPattern: string(LOGFMT_TS_LEVEL),
		}
		parsed, err := patternfactory.Parse(LOGFMT_TS_LEVEL, log)

		if err != nil {
			t.Error(err)
		}
		equal := reflect.DeepEqual(expected, parsed)
		if !equal {
			t.Errorf("Expected %+v but got %+v", expected, parsed)
		}
	}

}

func TestParseLOGFMT_LEVEL_TS(t *testing.T) {
	{
		var log = "level=info ts=2023-01-01T21:30:41.21756675Z caller=engine.go:199 component=querier org_id=fake msg=\"executing query\" type=range query=\"{ingress=\\\"vector-docker\\\", stack=\\\"security\\\", task=\\\"keycloak\\\"} | json | line_format \\\"{{.service_name}}@{{.service_node_name}} [{{.log_levelEmoji}} {{.level}}] [{{.tags}}] {{.message}}\\\"\" length=30m0s step=20m0s"
		ts, _ := time.Parse(time.RFC3339, "2023-01-01T21:30:41.21756675Z")

		expected := ParseResult{
			TimeStamp:   ts,
			LogLevel:    "info",
			UsedPattern: string(LOGFMT_LEVEL_TS),
		}
		parsed, err := patternfactory.Parse(LOGFMT_LEVEL_TS, log)

		if err != nil {
			t.Error(err)
		}
		equal := reflect.DeepEqual(expected, parsed)
		if !equal {
			t.Errorf("Expected %+v but got %+v", expected, parsed)
		}
	}

	{
		var log = "level=debug ts=2023-01-01T22:17:03.218151741Z caller=logging.go:76 traceID=42b7ebdd77087b60 msg=\"POST /api/v1/push (200) 2.543513ms\""
		ts, _ := time.Parse(time.RFC3339, "2023-01-01T22:17:03.218151741Z")

		expected := ParseResult{
			TimeStamp:   ts,
			LogLevel:    "debug",
			UsedPattern: string(LOGFMT_LEVEL_TS),
		}
		parsed, err := patternfactory.Parse(LOGFMT_LEVEL_TS, log)

		if err != nil {
			t.Error(err)
		}
		equal := reflect.DeepEqual(expected, parsed)
		if !equal {
			t.Errorf("Expected %+v but got %+v", expected, parsed)
		}
	}

}

func TestParseCONNECT_LOG(t *testing.T) {
	{
		var log = "[2023-01-01 18:20:55.198][14][warning][conn_handler] [source/server/active_stream_listener_base.cc:120] [C45161] adding to cleanup list"
		ts, _ := time.Parse(TimeFormatConsulConnect, "2023-01-01 18:20:55.198")

		expected := ParseResult{
			TimeStamp:   ts,
			LogLevel:    "warning",
			UsedPattern: string(CONNECT_LOG),
		}
		parsed, err := patternfactory.Parse(CONNECT_LOG, log)

		if err != nil {
			t.Error(err)
		}
		equal := reflect.DeepEqual(expected, parsed)
		if !equal {
			t.Errorf("Expected %+v but got %+v", expected, parsed)
		}
	}

}

func TestEcsAggTags(t *testing.T) {
	entry := &model.EcsLogEntry{}

	entry.Tags = append(entry.Tags, "1")
	entry.Tags = append(entry.Tags, "2")
	entry.Tags = append(entry.Tags, "3")
	if len(entry.Tags) != 3 {
		t.Errorf("Expect 3 elements bu got %d", len(entry.Tags))
	}
}
