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
		ts, err := time.Parse(time.RFC3339, "2022-12-08T12:21:02.594Z")
		if err != nil {
			t.Error(err)
		}
		expected := ParseResult{
			LogLevel:    model.LogLevel_error,
			TimeStamp:   ts,
			UsedPattern: CommonPattern.Name,
		}
		parsed, err := patternfactory.Parse("test", CommonPattern, log)

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
		ts, err := time.Parse(time.RFC3339, "2022-12-08T12:21:02.594Z")
		if err != nil {
			t.Error(err)
		}
		expected := ParseResult{
			LogLevel:    model.LogLevel_error,
			TimeStamp:   ts,
			UsedPattern: CommonPattern.Name,
		}
		parsed, err := patternfactory.Parse("test", CommonPattern, log)

		if err != nil {
			t.Error(err)
		}
		if parsed.TimeStamp.IsZero() {
			t.Errorf("Timestamp invalid  %s", parsed.TimeStamp.String())
		}
		equal := reflect.DeepEqual(expected, parsed)
		if !equal {
			t.Errorf("Expected %+v but got %+v", expected, parsed)
		}
	}
	{
		var log = "2022-12-08T12:21:02.594Z \"ERROR\" nomad.autopilot: 😜 Failed\nto reconcile current state with the desired state\nthird line mf\n1\n3"
		ts, err := time.Parse(time.RFC3339, "2022-12-08T12:21:02.594Z")
		if err != nil {
			t.Error(err)
		}
		expected := ParseResult{
			LogLevel:    model.LogLevel_error,
			TimeStamp:   ts,
			UsedPattern: CommonPattern.Name,
		}
		parsed, err := patternfactory.Parse("test", CommonPattern, log)

		if err != nil {
			t.Error(err)
		}
		equal := reflect.DeepEqual(expected, parsed)
		if parsed.TimeStamp.IsZero() {
			t.Errorf("Timestamp invalid  %s", parsed.TimeStamp.String())
		}
		if !equal {
			t.Errorf("Expected %+v but got %+v", expected, parsed)
		}
	}

	//logger.Info().Msgf("%s", parsed)
}

func TestParseTsKeyCloakStyle(t *testing.T) {
	{
		var log = "2023-01-01 21:22:59,772 ERROR [org.keycloak.services] (main) KC-SERVICES0010: Failed to add user 'admin' to realm 'master': user with username exists"
		//var log = "2023-01-01 21:22:59,772 ERROR [org.keycloak.services] (main) KC-SERVICES0010: Failed to add user 'admin' to realm 'master': user with username exists"
		ts, err := time.Parse(TimeFormatKeyCloak, "2023-01-01 21:22:59,772")
		if err != nil {
			t.Error(err)
		}
		expected := ParseResult{
			LogLevel:    model.LogLevel_error,
			TimeStamp:   ts,
			UsedPattern: KeyCloakPattern.Name,
		}
		parsed, err := patternfactory.Parse("test", KeyCloakPattern, log)

		if err != nil {
			t.Error(err)
		}
		equal := reflect.DeepEqual(expected, parsed)
		if parsed.TimeStamp.IsZero() {
			t.Errorf("Timestamp invalid  %s", parsed.TimeStamp.String())
		}
		if !equal {
			t.Errorf("Expected %+v but got %+v", expected, parsed)
		}
	}
}
func TestParseLOGFMT_TS_LEVEL_MSG(t *testing.T) {
	{
		var log = "time=\"2022-12-31T15:55:54.762121247Z\" level=warning msg=\"got error while decoding json\" error=\"unexpected EOF\" retries=1"
		ts, err := time.Parse(time.RFC3339, "2022-12-31T15:55:54.762121247Z")
		if err != nil {
			t.Error(err)
		}
		expected := ParseResult{
			TimeStamp:   ts,
			LogLevel:    model.LogLevel_warn,
			UsedPattern: CommonPattern.Name,
		}
		parsed, err := patternfactory.Parse("test", CommonPattern, log)

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
		ts, err := time.Parse(time.RFC3339, "2022-12-31T15:55:54.762121247Z")
		if err != nil {
			t.Error(err)
		}
		expected := ParseResult{
			TimeStamp:   ts,
			LogLevel:    model.LogLevel_warn,
			UsedPattern: CommonPattern.Name,
		}
		parsed, err := patternfactory.Parse("test", CommonPattern, log)

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
		ts, err := time.Parse(time.RFC3339, "2022-12-31T15:55:54.762121247Z")
		if err != nil {
			t.Error(err)
		}
		expected := ParseResult{
			TimeStamp:   ts,
			LogLevel:    model.LogLevel_warn,
			UsedPattern: CommonPattern.Name,
		}
		parsed, err := patternfactory.Parse("test", CommonPattern, log)

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
		ts, err := time.Parse(time.RFC3339, "2023-01-01T21:44:41.706702241Z")
		if err != nil {
			t.Error(err)
		}
		expected := ParseResult{
			TimeStamp:   ts,
			LogLevel:    model.LogLevel_debug,
			UsedPattern: CommonPattern.Name,
		}
		parsed, err := patternfactory.Parse("test", CommonPattern, log)

		if err != nil {
			t.Error(err)
		}
		equal := reflect.DeepEqual(expected, parsed)
		if !equal {
			t.Errorf("Expected %+v but got %+v", expected, parsed)
		}
	}
	{
		var log = "ts=2023-01-01T21:44:41.706702241Z er level=debug event=\"complete commit\" commitDuration=82.097µs"
		ts, err := time.Parse(time.RFC3339, "2023-01-01T21:44:41.706702241Z")
		if err != nil {
			t.Error(err)
		}
		expected := ParseResult{
			TimeStamp:   ts,
			LogLevel:    model.LogLevel_debug,
			UsedPattern: CommonPattern.Name,
		}
		parsed, err := patternfactory.Parse("test", CommonPattern, log)

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
		ts, err := time.Parse(time.RFC3339, "2023-01-01T22:14:36.47634233Z")
		if err != nil {
			t.Error(err)
		}
		expected := ParseResult{
			LogLevel:    model.LogLevel_debug,
			TimeStamp:   ts,
			UsedPattern: CommonPattern.Name,
		}
		parsed, err := patternfactory.Parse("test", CommonPattern, log)

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
		ts, err := time.Parse(time.RFC3339, "2023-01-01T22:16:44.602022589Z")
		if err != nil {
			t.Error(err)
		}
		expected := ParseResult{
			TimeStamp:   ts,
			LogLevel:    model.LogLevel_info,
			UsedPattern: CommonPattern.Name,
		}
		parsed, err := patternfactory.Parse("test", CommonPattern, log)

		if err != nil {
			t.Error(err)
		}
		if parsed.TimeStamp.IsZero() {
			t.Errorf("Timestamp invalid  %s", parsed.TimeStamp.String())
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
		ts, err := time.Parse(time.RFC3339, "2023-01-01T21:30:41.21756675Z")
		if err != nil {
			t.Error(err)
		}
		expected := ParseResult{
			TimeStamp:   ts,
			LogLevel:    model.LogLevel_info,
			UsedPattern: CommonPattern.Name,
		}
		parsed, err := patternfactory.Parse("test", CommonPattern, log)

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
		ts, err := time.Parse(time.RFC3339, "2023-01-01T22:17:03.218151741Z")
		if err != nil {
			t.Error(err)
		}
		expected := ParseResult{
			TimeStamp:   ts,
			LogLevel:    model.LogLevel_debug,
			UsedPattern: CommonPattern.Name,
		}
		parsed, err := patternfactory.Parse("test", CommonPattern, log)

		if err != nil {
			t.Error(err)
		}
		if parsed.TimeStamp.IsZero() {
			t.Errorf("Timestamp invalid  %s", parsed.TimeStamp.String())
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
		ts, err := time.Parse(TimeFormatConsulConnect, "2023-01-01 18:20:55.198")
		if err != nil {
			t.Error(err)
		}
		expected := ParseResult{
			TimeStamp:   ts,
			LogLevel:    model.LogLevel_warn,
			UsedPattern: ConsulConnectPattern.Name,
		}
		parsed, err := patternfactory.Parse("test", ConsulConnectPattern, log)

		if err != nil {
			t.Error(err)
		}
		if parsed.TimeStamp.IsZero() {
			t.Errorf("Timestamp invalid  %s", parsed.TimeStamp.String())
		}
		equal := reflect.DeepEqual(expected, parsed)
		if !equal {
			t.Errorf("Expected %+v but got %+v", expected, parsed)
		}
	}

}

func TestMimirLogs(t *testing.T) {
	{
		var log = "level=debug ts=2023-01-02T10:02:56.282190964Z caller=broadcast.go:48 msg=\"Invalidating forwarded broadcast\" key=collectors/ring version=85 oldVersion=84 content=[777c9daf009c] oldContent=[777c9daf009c]"
		ts, err := time.Parse(time.RFC3339, "2023-01-02T10:02:56.282190964Z")
		if err != nil {
			t.Error(err)
		}
		expected := ParseResult{
			TimeStamp:   ts,
			LogLevel:    model.LogLevel_debug,
			UsedPattern: CommonPattern.Name,
		}
		parsed, err := patternfactory.Parse("test", CommonPattern, log)

		if err != nil {
			t.Error(err)
		}
		if parsed.TimeStamp.IsZero() {
			t.Errorf("Timestamp invalid  %s", parsed.TimeStamp.String())
		}
		equal := reflect.DeepEqual(expected, parsed)
		if !equal {
			t.Errorf("Expected %+v but got %+v", expected, parsed)
		}
	}

	{
		var log = "ts=2023-01-02T10:02:56.085691108Z caller=spanlogger.go:80 user=anonymous level=debug event=\"start commit\" succeededSamplesCount=500 failedSamplesCount=0 succeededExemplarsCount=0 failedExemplarsCount=0]"
		ts, err := time.Parse(time.RFC3339, "2023-01-02T10:02:56.085691108Z")
		if err != nil {
			t.Error(err)
		}
		expected := ParseResult{
			TimeStamp:   ts,
			LogLevel:    model.LogLevel_debug,
			UsedPattern: CommonPattern.Name,
		}
		parsed, err := patternfactory.Parse("test", CommonPattern, log)

		if err != nil {
			t.Error(err)
		}
		if parsed.TimeStamp.IsZero() {
			t.Errorf("Timestamp invalid  %s", parsed.TimeStamp.String())
		}
		equal := reflect.DeepEqual(expected, parsed)
		if !equal {
			t.Errorf("Expected %+v but got %+v", expected, parsed)
		}
	}
}

func TestTempoLogs(t *testing.T) {
	{
		var log = "ts=2023-01-02T14:01:42.769978691Z caller=dedupe.go:112 tenant=single-tenant component=remote level=warn remote_name=3b8ad9 url=http:///api/v1/push msg=\"Failed to send batch, retrying\" err=\"Post \\\"http:///api/v1/push\\\": http: no Host in request URL\""
		ts, err := time.Parse(time.RFC3339, "2023-01-02T14:01:42.769978691Z")
		if err != nil {
			t.Error(err)
		}
		expected := ParseResult{
			TimeStamp:   ts,
			LogLevel:    model.LogLevel_warn,
			UsedPattern: CommonPattern.Name,
		}
		parsed, err := patternfactory.Parse("test", CommonPattern, log)

		if err != nil {
			t.Error(err)
		}
		if parsed.TimeStamp.IsZero() {
			t.Errorf("Timestamp invalid  %s", parsed.TimeStamp.String())
		}
		equal := reflect.DeepEqual(expected, parsed)
		if !equal {
			t.Errorf("Expected %+v but got %+v", expected, parsed)
		}
	}

}

func TestConnectProxLogs(t *testing.T) {
	{
		var log = "[2023-01-02 12:49:46.745][1][debug][upstream] [source/common/upstream/strict_dns_cluster.cc:178] DNS refresh rate reset for tempo-zipkin.service.consul, refresh rate 5000 ms/ring version=85 oldVersion=84 content=[777c9daf009c] oldContent=[777c9daf009c]"
		ts, err := time.Parse(TimeFormatConsulConnect, "2023-01-02 12:49:46.745")
		if err != nil {
			t.Error(err)
		}
		expected := ParseResult{
			TimeStamp:   ts,
			LogLevel:    model.LogLevel_debug,
			UsedPattern: ConsulConnectPattern.Name,
		}
		parsed, err := patternfactory.Parse("test", ConsulConnectPattern, log)

		if err != nil {
			t.Error(err)
		}
		if parsed.TimeStamp.IsZero() {
			t.Errorf("Timestamp invalid  %s", parsed.TimeStamp.String())
		}
		equal := reflect.DeepEqual(expected, parsed)
		if !equal {
			t.Errorf("Expected %+v but got %+v", expected, parsed)
		}
	}

}
func TestPostgresLogs(t *testing.T) {
	{
		var log = "2023-01-02 13:22:41.280 UTC [10838] FATAL:  role \"root\" does not exist] and found ts is [2023-01-02 13:22:41.280"
		ts, err := time.Parse(TimeformatCommonUTC, "2023-01-02 13:22:41.280 UTC")
		if err != nil {
			t.Error(err)
		}
		expected := ParseResult{
			TimeStamp:   ts,
			LogLevel:    model.LogLevel_fatal,
			UsedPattern: CommonUtcPattern.Name,
		}
		parsed, err := patternfactory.Parse("test", CommonUtcPattern, log)

		if err != nil {
			t.Error(err)
		}
		if parsed.TimeStamp.IsZero() {
			t.Errorf("Timestamp invalid  %s", parsed.TimeStamp.String())
		}
		equal := reflect.DeepEqual(expected, parsed)
		if !equal {
			t.Errorf("Expected %+v but got %+v", expected, parsed)
		}
	}

}

func TestNexusLogs(t *testing.T) {
	{
		var log = "2023-01-02 14:00:00,003+0000 INFO  [quartz-10-thread-20] *SYSTEM org.sonatype.nexus.quartz.internal.task.QuartzTaskInfo - Task 'Storage facet cleanup' [repository.storage-facet-cle\nanup] state change WAITING -> RUNNING"
		ts, err := time.Parse(TimeFormatUtcCommaSecondAndTs, "2023-01-02 14:00:00,003+0000")
		if err != nil {
			t.Error(err)
		}
		expected := ParseResult{
			TimeStamp:   ts,
			LogLevel:    model.LogLevel_info,
			UsedPattern: CommonUtcPatternWithCommaTsAndTz.Name,
		}
		parsed, err := patternfactory.Parse("test", CommonUtcPatternWithCommaTsAndTz, log)

		if err != nil {
			t.Error(err)
		}
		if parsed.TimeStamp.IsZero() {
			t.Errorf("Timestamp invalid  %s", parsed.TimeStamp.String())
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
