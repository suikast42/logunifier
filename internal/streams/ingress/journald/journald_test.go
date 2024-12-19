package journald

import (
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"github.com/suikast42/logunifier/internal/config"
	"github.com/suikast42/logunifier/pkg/model"
	"github.com/suikast42/logunifier/pkg/patterns"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

var (
	patternfactory *patterns.PatternFactory
	logger         zerolog.Logger
)

func init() {
	logger = config.Logger()
	_, err := patterns.Initialize()
	if err != nil {
		logger.Error().Err(err).Stack().Msg("Can't initialize pattern factory")
		os.Exit(1)
	}
	patternfactory = patterns.Instance()
}
func TestDefaultPatterContainerLogTest(t *testing.T) {
	expectedMessage := "TestMessage"
	log := TestMetaLogFromJournalD([]byte(testJournaldContainerLog), expectedMessage, t)

	parse := patternfactory.Parse(log)

	if parse.Message != "TestMessage" {
		t.Errorf("Expected %+v but got %+v", expectedMessage, parse.Message)
	}

}

func TestDockerServiceLog(t *testing.T) {
	log := TestMetaLogFromJournalDFromConst([]byte(testJournaldDockerServiceLog), t)
	parsed := patternfactory.Parse(log)
	if parsed == nil {
		t.Error("Expected not nil but got nil")
	}
	if parsed.Log == nil {
		t.Error("Expected not nil but got nil")
	}

	grok := parsed.Log.PatternKey

	if grok != model.MetaLog_LogFmt.String() {
		t.Errorf("Expected pattern %s but got %s", model.MetaLog_LogFmt.String(), grok)
	}

	if parsed.Log.Level != model.LogLevel_error {
		t.Errorf("Expected Log level %+v but got %+v", model.LogLevel_error, parsed.Log.Level)
	}

	if parsed.Log.LevelEmoji != model.LogLevelToEmoji(model.LogLevel_error) {
		t.Errorf("Expected Log level emoji  %+v but got %+v", model.LogLevelToEmoji(model.LogLevel_error), model.LogLevelToEmoji(parsed.Log.Level))
	}
	if !strings.Contains(parsed.Message, "collecting stats for") {
		t.Errorf("Expected Log message starts with  %+v but got %+v", "collecting stats for", parsed.Message)
	}

	if len(parsed.ProcessError.Reason) > 0 {
		t.Errorf("Expected no parse errors but got %+v", parsed.ProcessError)
	}
	//2023-03-10T18:53:52
	time := parsed.GetTimeStamp()
	year := 2023
	month := "March"
	day := 10
	hour := 18
	minute := 53
	second := 52
	if time.IsZero() {
		t.Errorf("Expected time not zero  but got %+v", time)
	}
	if time.Year() != year {
		t.Errorf("Expected %d  but got %d", year, time.Year())
	}
	if time.Month().String() != "March" {
		t.Errorf("Expected %s  but got %s", month, time.Month())
	}
	if time.Day() != day {
		t.Errorf("Expected %d  but got %d", day, time.Day())
	}
	if time.Minute() != minute {
		t.Errorf("Expected %d  but got %d", minute, time.Minute())
	}
	if time.Hour() != hour {
		t.Errorf("Expected %d  but got %d", hour, time.Hour())
	}
	if time.Second() != second {
		t.Errorf("Expected %d  but got %d", second, time.Second())
	}

}

func TestGrafanaLog(t *testing.T) {
	log := TestMetaLogFromJournalDFromConst([]byte(testJournaldGrafanaLog), t)
	parsed := patternfactory.Parse(log)
	if parsed == nil {
		t.Error("Expected not nil but got nil")
	}

	if parsed.Log == nil {
		t.Error("Expected not nil but got nil")
	}

	grok := parsed.Log.PatternKey

	if grok != model.MetaLog_LogFmt.String() {
		t.Errorf("Expected pattern [%s] but got [%s]", model.MetaLog_LogFmt.String(), grok)
	}

	if parsed.Log.Level != model.LogLevel_info {
		t.Errorf("Expected Log level %+v but got %+v", model.LogLevel_info, parsed.Log.Level)
	}

	if parsed.Log.LevelEmoji != model.LogLevelToEmoji(model.LogLevel_info) {
		t.Errorf("Expected Log level emoji  %+v but got %+v", model.LogLevelToEmoji(model.LogLevel_info), model.LogLevelToEmoji(parsed.Log.Level))
	}
	if !strings.Contains(parsed.Message, "Initialized channel handler") {
		t.Errorf("Expected Log message starts with  %+v but got %+v", "Initialized channel handler", parsed.Message)
	}

	if len(parsed.ProcessError.Reason) > 0 {
		t.Errorf("Expected no parse errors but got %+v", parsed.ProcessError)
	}
	//2023-03-16T20:43:56.274825539Z
	time := parsed.GetTimeStamp()
	year := 2023
	month := "March"
	day := 16
	hour := 20
	minute := 43
	second := 56
	if time.IsZero() {
		t.Errorf("Expected time not zero  but got %+v", time)
	}
	if time.Year() != year {
		t.Errorf("Expected %d  but got %d", year, time.Year())
	}
	if time.Month().String() != "March" {
		t.Errorf("Expected %s  but got %s", month, time.Month())
	}
	if time.Day() != day {
		t.Errorf("Expected %d  but got %d", day, time.Day())
	}
	if time.Minute() != minute {
		t.Errorf("Expected %d  but got %d", minute, time.Minute())
	}
	if time.Hour() != hour {
		t.Errorf("Expected %d  but got %d", hour, time.Hour())
	}
	if time.Second() != second {
		t.Errorf("Expected %d  but got %d", second, time.Second())
	}

}

func TestTraefikInvalidLogFmt(t *testing.T) {
	log := TestMetaLogFromJournalDFromConst([]byte(testJournaldTraefikInvalidLogFmt), t)
	parsed := patternfactory.Parse(log)
	if parsed == nil {
		t.Error("Expected not nil but got nil")
	}
	if parsed.Log == nil {
		t.Error("Expected not nil but got nil")
	}

	grok := parsed.Log.PatternKey

	if grok != model.MetaLog_LogFmt.String() {
		t.Errorf("Expected pattern [%s] but got [%s]", model.MetaLog_LogFmt.String(), grok)
	}

	if len(parsed.ProcessError.Reason) == 0 {
		t.Error("Expect a process error bu got nothing")
	}

	if len(parsed.Message) == 0 {
		t.Error("Expected a message nothing")
	}

}

func TestNomadLog(t *testing.T) {
	log := TestMetaLogFromJournalDFromConst([]byte(testJournaldNomadLog), t)
	parsed := patternfactory.Parse(log)
	if parsed == nil {
		t.Error("Expected not nil but got nil")
	}
	if parsed.Log == nil {
		t.Error("Expected not nil but got nil")
	}

	grok := parsed.Log.PatternKey
	if grok != model.MetaLog_TsLevelMsg.String() {
		t.Errorf("Expected pattern [%s] but got [%s]", model.MetaLog_TsLevelMsg.String(), grok)
	}

	if len(parsed.ProcessError.Reason) > 0 {
		t.Errorf("Should not contain a process error. But contains %s", parsed.ProcessError.Reason)
	}

	if len(parsed.Message) == 0 {
		t.Error("Expected a message nothing")
	}

	if parsed.Log.Level != model.LogLevel_debug {
		t.Errorf("Expected a debug level but got %+v", parsed.Log.Level)
	}
	if parsed.GetTimeStamp().IsZero() {
		t.Errorf("Expected a valid ts but got %+v", parsed.GetTimeStamp())
	}

	//2023-03-20T15:06:45.057Z
	time := parsed.GetTimeStamp()
	year := 2023
	month := "March"
	day := 20
	hour := 15
	minute := 06
	second := 45
	if time.Year() != year {
		t.Errorf("Expected %d  but got %d", year, time.Year())
	}
	if time.Month().String() != "March" {
		t.Errorf("Expected %s  but got %s", month, time.Month())
	}
	if time.Day() != day {
		t.Errorf("Expected %d  but got %d", day, time.Day())
	}
	if time.Minute() != minute {
		t.Errorf("Expected %d  but got %d", minute, time.Minute())
	}
	if time.Hour() != hour {
		t.Errorf("Expected %d  but got %d", hour, time.Hour())
	}
	if time.Second() != second {
		t.Errorf("Expected %d  but got %d", second, time.Second())
	}
}

func TestConsulConnectLog(t *testing.T) {
	log := TestMetaLogFromJournalDFromConst([]byte(testJournaldConsulConnect), t)
	parsed := patternfactory.Parse(log)
	if parsed == nil {
		t.Error("Expected not nil but got nil")
	}
	if parsed.Log == nil {
		t.Error("Expected not nil but got nil")
	}

	grok := parsed.Log.PatternKey

	if grok != model.MetaLog_Envoy.String() {
		t.Errorf("Expected pattern [%s] but got [%s]", model.MetaLog_Envoy.String(), grok)
	}

	if len(parsed.ProcessError.Reason) > 0 {
		t.Errorf("Should not contain a process error. But contains %s", parsed.ProcessError.Reason)
	}

	if len(parsed.Message) == 0 {
		t.Error("Expected a message nothing")
	}

	if parsed.Log.Level != model.LogLevel_debug {
		t.Errorf("Expected a debug level but got %+v", parsed.Log.Level)
	}
	if parsed.GetTimeStamp().IsZero() {
		t.Errorf("Expected a valid ts but got %+v", parsed.GetTimeStamp())
	}

	//2023-03-30 12:33:20.424
	time := parsed.GetTimeStamp()
	year := 2023
	month := "March"
	day := 30
	hour := 12
	minute := 33
	second := 20
	if time.Year() != year {
		t.Errorf("Expected %d  but got %d", year, time.Year())
	}
	if time.Month().String() != "March" {
		t.Errorf("Expected %s  but got %s", month, time.Month())
	}
	if time.Day() != day {
		t.Errorf("Expected %d  but got %d", day, time.Day())
	}
	if time.Minute() != minute {
		t.Errorf("Expected %d  but got %d", minute, time.Minute())
	}
	if time.Hour() != hour {
		t.Errorf("Expected %d  but got %d", hour, time.Hour())
	}
	if time.Second() != second {
		t.Errorf("Expected %d  but got %d", second, time.Second())
	}
}

func TestInvalidTsLevelMsg(t *testing.T) {
	log := TestMetaLogFromJournalDFromConst([]byte(testJournaldInvalidTsLevelMsg), t)
	parsed := patternfactory.Parse(log)
	if parsed == nil {
		t.Error("Expected not nil but got nil")
	}
	if parsed.Log == nil {
		t.Error("Expected not nil but got nil")
	}

	grok := parsed.Log.PatternKey

	if grok != model.MetaLog_TsLevelMsg.String() {
		t.Errorf("Expected pattern [%s] but got [%s]", model.MetaLog_TsLevelMsg.String(), grok)
	}

	if !parsed.HasProcessError() {
		t.Error("Expect process error bu got nothing ")
	}

	if len(parsed.Message) == 0 {
		t.Error("Expected a message nothing")
	}

	if parsed.Log.Level == model.LogLevel_not_set {
		t.Errorf("Expected a %+v level but got %+v", model.LogLevel_not_set, parsed.Log.Level)
	}
	if parsed.GetTimeStamp().IsZero() {
		t.Errorf("Expected a valid ts but got %+v", parsed.GetTimeStamp())
	}

	if parsed.Message != "Invalid message" {
		t.Errorf("Expected message [Invalid message] but got [%s]", parsed.Message)
	}

}

func TestLogunifier(t *testing.T) {
	log := TestMetaLogFromJournalDFromConst([]byte(testJournaldLogunifier), t)
	parsed := patternfactory.Parse(log)
	if parsed == nil {
		t.Error("Expected not nil but got nil")
	}
	if parsed.Log == nil {
		t.Error("Expected not nil but got nil")
	}

	grok := parsed.Log.PatternKey

	if grok != model.MetaLog_TsLevelMsg.String() {
		t.Errorf("Expected pattern [%s] but got [%s]", model.MetaLog_TsLevelMsg.String(), grok)
	}

	if len(parsed.ProcessError.Reason) > 0 {
		t.Errorf("Should not contain a process error. But contains %s", parsed.ProcessError.Reason)
	}

	if len(parsed.Message) == 0 {
		t.Error("Expected a message nothing")
	}

	if parsed.Log.Level != model.LogLevel_debug {
		t.Errorf("Expected a debug level but got %+v", parsed.Log.Level)
	}
	if parsed.GetTimeStamp().IsZero() {
		t.Errorf("Expected a valid ts but got %+v", parsed.GetTimeStamp())
	}

	//2023-03-30T20:13:52.774125Z
	time := parsed.GetTimeStamp()
	year := 2023
	month := "March"
	day := 30
	hour := 20
	minute := 13
	second := 52
	if time.Year() != year {
		t.Errorf("Expected %d  but got %d", year, time.Year())
	}
	if time.Month().String() != "March" {
		t.Errorf("Expected %s  but got %s", month, time.Month())
	}
	if time.Day() != day {
		t.Errorf("Expected %d  but got %d", day, time.Day())
	}
	if time.Minute() != minute {
		t.Errorf("Expected %d  but got %d", minute, time.Minute())
	}
	if time.Hour() != hour {
		t.Errorf("Expected %d  but got %d", hour, time.Hour())
	}
	if time.Second() != second {
		t.Errorf("Expected %d  but got %d", second, time.Second())
	}

	if parsed.Message != "Nothing to validate after 10s " {
		t.Errorf("Expected message [Nothing to validate after 10s] but got [%s]", parsed.Message)
	}

}

func TestTraefikLogLogFmt(t *testing.T) {
	log := TestMetaLogFromJournalDFromConst([]byte(testTraefikLogLogfmt), t)
	if log.HasProcessErrors() {
		t.Errorf("Metalog with process errors %s", log.EcsLogEntry.ProcessError.Reason)
		t.Errorf("Raw data [%s]", log.EcsLogEntry.ProcessError.RawData)
	}
	parsedEcs := patternfactory.Parse(log)

	if parsedEcs.HasProcessError() {
		t.Error("Expect no process error bu got " + parsedEcs.ProcessError.Reason)
	}

	if log.PatternKey != model.MetaLog_Traefik {
		t.Errorf("Expected pattern [%s] but got [%s]", model.MetaLog_Traefik, log.PatternKey)
	}
	if parsedEcs.Log == nil {
		t.Error("Expect message but was nil")
	}

	if parsedEcs.Log.Origin == nil {
		t.Error("Expect origin but was nil")
	}

	if parsedEcs.Log.Level != model.LogLevel_debug {
		t.Errorf("Expect level debug but got %s", parsedEcs.Log.Level)
	}

}

func TestEcsOverJournald(t *testing.T) {
	log := TestMetaLogFromJournalDFromConst([]byte(testJournaldEcs), t)

	if log.HasProcessErrors() {
		t.Errorf("Metalog with process errors %s", log.EcsLogEntry.ProcessError.Reason)
		t.Errorf("Raw data [%s]", log.EcsLogEntry.ProcessError.RawData)
	}
	parsedEcs := patternfactory.Parse(log)

	if parsedEcs == nil {
		t.Error("Expected not nil but got nil")
	}

	if parsedEcs.Log == nil {
		t.Error("Expected not nil but got nil")
	}

	grok := parsedEcs.Log.PatternKey

	if grok != model.MetaLog_Ecs.String() {
		t.Errorf("Expected pattern [%s] but got [%s]", model.MetaLog_Ecs.String(), grok)
	}

	if len(parsedEcs.ProcessError.Reason) > 0 {
		t.Errorf("Should not contain a process error. But contains %s", parsedEcs.ProcessError.Reason)
	}

	if len(parsedEcs.Message) == 0 {
		t.Error("Expected a message nothing")
	}

	if parsedEcs.Log.Level != model.LogLevel_debug {
		t.Errorf("Expected a debug level but got %+v", parsedEcs.Log.Level)
	}
	if parsedEcs.GetTimeStamp().IsZero() {
		t.Errorf("Expected a valid ts but got %+v", parsedEcs.GetTimeStamp())
	}

	if parsedEcs.Container == nil {
		t.Errorf("Expected Container metadata is nil")
	}

	if parsedEcs.Container != nil && !strings.HasPrefix(parsedEcs.Container.Name, "grafana") {
		t.Errorf("Expected Container name starts with grafana but is %s ", parsedEcs.Container.Name)
	}
	if parsedEcs.Log.Ingress != "test" {
		t.Errorf("Expected ingress subject test but is %s", parsedEcs.Log.Ingress)
	}

	if parsedEcs.Service == nil {
		t.Errorf("Expected Service metadata is nil")
	}

	if parsedEcs.Service != nil && parsedEcs.Service.Stack != "observability" {
		t.Errorf("Expected Service stack is observability but was %s", parsedEcs.Service.Stack)
	}

	if parsedEcs.Service != nil && parsedEcs.Service.Version != "9.4.3.0" {
		t.Errorf("Expected Service version is 9.4.3.0 but was %s", parsedEcs.Service.Version)
	}

	if parsedEcs.Service != nil && parsedEcs.Service.Group != "grafana" {
		t.Errorf("Expected Service version is grafana but was %s", parsedEcs.Service.Group)
	}

	if parsedEcs.Host == nil {
		t.Errorf("Expected Host metadata is nil")
	}

	if parsedEcs.Host != nil && parsedEcs.Host.Name != "worker-01" {
		t.Errorf("Expected Host name is worker-01 but was %s", parsedEcs.Host.Name)
	}

	if parsedEcs.Host != nil && parsedEcs.Host.Id != "ceacb99587e34bcc840bc7a7cc0d4453" {
		t.Errorf("Expected Host is is ceacb99587e34bcc840bc7a7cc0d4453 but was %s", parsedEcs.Host.Id)
	}
	//2023-03-30T20:13:52.774125Z
	time := parsedEcs.GetTimeStamp()
	year := 2023
	month := "June"
	day := 7
	hour := 13
	minute := 8
	second := 51
	if time.Year() != year {
		t.Errorf("Expected %d  but got %d", year, time.Year())
	}
	if time.Month().String() != "June" {
		t.Errorf("Expected %s  but got %s", month, time.Month())
	}
	if time.Day() != day {
		t.Errorf("Expected %d  but got %d", day, time.Day())
	}
	if time.Minute() != minute {
		t.Errorf("Expected %d  but got %d", minute, time.Minute())
	}
	if time.Hour() != hour {
		t.Errorf("Expected %d  but got %d", hour, time.Hour())
	}
	if time.Second() != second {
		t.Errorf("Expected %d  but got %d", second, time.Second())
	}

	if len(parsedEcs.Message) == 0 {
		t.Error("Expected a message nothing")
	} else {
		if parsedEcs.Log.Level != model.LogLevel_debug {
			t.Errorf("Expected a debug level but got %+v", parsedEcs.Log.Level)
		}
		if !strings.EqualFold(parsedEcs.Log.Logger, "com.boxbay.wms.internal.test.curd.WmsCrudTest") {
			t.Errorf("Expected a logger named com.boxbay.wms.internal.test.curd.WmsCrudTest got [%v]", parsedEcs.Log.Logger)
		}
	}

}

func TestNativeEcsFromJson(t *testing.T) {

	parsedEcs := model.EcsLogEntry{}
	err := parsedEcs.FromJson([]byte(testNatviceEcs))
	if err != nil {
		t.Error("Can't unmarshal ecs ", err)
	}

	time := parsedEcs.GetTimeStamp()
	year := 2023
	month := "June"
	day := 7
	hour := 13
	minute := 8
	second := 51

	if time.IsZero() {
		t.Errorf("Expected time not zero  but got %+v", time)
	}
	if time.Year() != year {
		t.Errorf("Expected %d  but got %d", year, time.Year())
	}
	if time.Month().String() != "June" {
		t.Errorf("Expected %s but got %s", month, time.Month())
	}
	if time.Day() != day {
		t.Errorf("Expected %d  but got %d", day, time.Day())
	}
	if time.Minute() != minute {
		t.Errorf("Expected %d  but got %d", minute, time.Minute())
	}
	if time.Hour() != hour {
		t.Errorf("Expected %d  but got %d", hour, time.Hour())
	}
	if time.Second() != second {
		t.Errorf("Expected %d  but got %d", second, time.Second())
	}

	if !strings.Contains(parsedEcs.Message, "Running with Spring Boot") {
		t.Errorf("Expected message [Running with Spring Boot] but got [%s]", parsedEcs.Message)
	}

	if len(parsedEcs.Message) == 0 {
		t.Error("Expected a message nothing")
	} else {
		if parsedEcs.Log.Level != model.LogLevel_debug {
			t.Errorf("Expected a debug level but got %+v", parsedEcs.Log.Level)
		}
		if !strings.EqualFold(parsedEcs.Log.Logger, "com.boxbay.wms.internal.test.curd.WmsCrudTest") {
			t.Errorf("Expected a logger named com.boxbay.wms.internal.test.curd.WmsCrudTest got [%v]", parsedEcs.Log.Logger)
		}
	}

}
func TestNativeEcsSerde(t *testing.T) {
	ts := time.Now().Add(-time.Hour * 360)
	message := "Test message"
	log := model.Log{
		File: &model.Log_File{
			Path: "/mnt/var/logs",
		},
		Level:      model.LogLevel_fatal,
		Logger:     "fooLogger",
		ThreadName: "fooThread",
		Origin: &model.Log_Origin{
			File: &model.Log_Origin_File{
				Line: "42",
				Name: "KannAlles",
			},
			Function: "hasiFunc",
		},
		Original: "",
		Syslog: &model.Log_Syslog{
			Facility: &model.Log_Syslog_Facility{
				Code: "0",
				Name: "Hans",
			},
			Priority: "0",
			Severity: &model.Log_Syslog_Severity{
				Code: "0",
				Name: "HasiSevirity",
			},
		},
		LevelEmoji: "",
	}
	service := model.Service{
		EphemeralId: "42",
		Id:          "42",
		Name:        "HasiService",
		Node:        &model.Service_Node{Name: "HasiNode"},
		State:       "Up",
		Type:        "Node",
		Version:     "1.2",
	}
	processError := model.ProcessError{
		Reason:  "Test",
		RawData: "{2321}",
		Subject: "hasi.foo.bongo",
	}
	toSerialize := &model.EcsLogEntry{}
	toSerialize.SetTimeStamp(timestamppb.New(ts))

	toSerialize.Message = message
	toSerialize.Log = &log
	toSerialize.Id = "Id42"
	toSerialize.Service = &service
	toSerialize.ProcessError = &processError

	json, err := toSerialize.ToJson()
	if err != nil {
		t.Errorf("Can't marshal ecs %s", err)
	}

	fromJson := &model.EcsLogEntry{}
	err = fromJson.FromJson(json)
	if err != nil {
		t.Errorf("Can't unmarshal ecs %s", err)
	}
	if !proto.Equal(toSerialize, fromJson) {
		t.Errorf("\ntoSerialize\n[%v]\nfromJson\n[%v]", toSerialize, fromJson)
	}

}

func TestMultiLine(t *testing.T) {
	part1 := `
  {
    "CONTAINER_ID_FULL": "fd973ef56c295e4cbfcee89144320326d4185ab5ddffbf4f8bd3cbb08b77f441",
    "_SYSTEMD_INVOCATION_ID": "c83ed454c19b47948dc6642f29322041",
    "MESSAGE": null,
    "COM_HASHICORP_NOMAD_ALLOC_ID": "3607011c-319e-9e2b-db2a-3ed3bc7f1f79",
    "_SYSTEMD_UNIT": "docker.service",
    "_CMDLINE": "/usr/bin/dockerd -H fd:// --containerd=/run/containerd/containerd.sock",
    "CONTAINER_TAG": "fd973ef56c29",
    "_TRANSPORT": "journal",
    "PRIORITY": "6",
    "COM_GITHUB_LOGUNIFIER_APPLICATION_NAME": "amovabi_services.devicetracking_service[0]",
    "COM_HASHICORP_NOMAD_NODE_NAME": "worker-01",
    "COM_GITHUB_LOGUNIFIER_APPLICATION_PATTERN_KEY": "ecs",
    "CONTAINER_ID": "fd973ef56c29",
    "SYSLOG_IDENTIFIER": "fd973ef56c29",
    "COM_GITHUB_LOGUNIFIER_APPLICATION_ENV": "dev",
    "_SOURCE_REALTIME_TIMESTAMP": "1724414678319480",
    "__REALTIME_TIMESTAMP": "1724414678321239",
    "__MONOTONIC_TIMESTAMP": "17895303600",
    "_SYSTEMD_CGROUP": "/system.slice/docker.service",
    "_BOOT_ID": "03ff567220e2431795e42279271b4808",
    "IMAGE_NAME": "registry.cloud.private/amovabi/device-tracking-service:2.0.0-CR23",
    "COM_HASHICORP_NOMAD_TASK_GROUP_NAME": "devicetracking_service",
    "CONTAINER_NAME": "devicetracking_service_task-3607011c-319e-9e2b-db2a-3ed3bc7f1f79",
    "_GID": "0",
    "COM_HASHICORP_NOMAD_JOB_ID": "amovabi_services",
    "CONTAINER_LOG_EPOCH": "6845af5b79aa83a96d0b213bc15da43cb0c1c2a75367bae29afc6533eae26875",
    "COM_HASHICORP_NOMAD_NAMESPACE": "default",
    "_MACHINE_ID": "f903ee60ec3c49c6b810aa51534b93c3",
    "_UID": "0",
    "_SYSTEMD_SLICE": "system.slice",
    "COM_HASHICORP_NOMAD_JOB_NAME": "amovabi_services",
    "CONTAINER_LOG_ORDINAL": "6378",
    "COM_HASHICORP_NOMAD_NODE_ID": "70f52bd2-411a-bc3e-6244-2883e42568e9",
    "COM_GITHUB_LOGUNIFIER_APPLICATION_VERSION": "2.0.0-CR23",
    "_EXE": "/usr/bin/dockerd",
    "COM_HASHICORP_NOMAD_TASK_NAME": "devicetracking_service_task",
    "__CURSOR": "s=52d6bbae8b5445f7a4155820202b5da6;i=7822d9;b=03ff567220e2431795e42279271b4808;m=42aa4a9b0;t=620589458c857;x=82ad08a70faa61e6",
    "SYSLOG_TIMESTAMP": "2024-08-23T12:04:38.319435796Z",
    "_SELINUX_CONTEXT": "unconfined\n",
    "COM_GITHUB_LOGUNIFIER_APPLICATION_ORG": "amova",
    "_HOSTNAME": "worker-01",
    "CONTAINER_PARTIAL_MESSAGE": "true",
    "CONTAINER_PARTIAL_LAST": "false",
    "CONTAINER_PARTIAL_ORDINAL": "1",
    "_COMM": "dockerd",
    "_CAP_EFFECTIVE": "1ffffffffff",
    "_PID": "1251",
    "CONTAINER_PARTIAL_ID": "0c4de2f5604ca894fe3896bbc0ecf9477ffbdea5967e53f6223fdeb68364f80c"
  }
`

	part2 := `
{
    "_SYSTEMD_CGROUP": "/system.slice/docker.service",
    "SYSLOG_IDENTIFIER": "fd973ef56c29",
    "_TRANSPORT": "journal",
    "_SYSTEMD_UNIT": "docker.service",
    "_SYSTEMD_INVOCATION_ID": "c83ed454c19b47948dc6642f29322041",
    "COM_GITHUB_LOGUNIFIER_APPLICATION_ORG": "amova",
    "_GID": "0",
    "CONTAINER_LOG_EPOCH": "6845af5b79aa83a96d0b213bc15da43cb0c1c2a75367bae29afc6533eae26875",
    "CONTAINER_LOG_ORDINAL": "6379",
    "COM_GITHUB_LOGUNIFIER_APPLICATION_VERSION": "2.0.0-CR23",
    "_MACHINE_ID": "f903ee60ec3c49c6b810aa51534b93c3",
    "COM_HASHICORP_NOMAD_NODE_ID": "70f52bd2-411a-bc3e-6244-2883e42568e9",
    "CONTAINER_PARTIAL_LAST": "true",
    "CONTAINER_NAME": "devicetracking_service_task-3607011c-319e-9e2b-db2a-3ed3bc7f1f79",
    "COM_HASHICORP_NOMAD_TASK_NAME": "devicetracking_service_task",
    "_COMM": "dockerd",
    "_CMDLINE": "/usr/bin/dockerd -H fd:// --containerd=/run/containerd/containerd.sock",
    "CONTAINER_TAG": "fd973ef56c29",
    "_BOOT_ID": "03ff567220e2431795e42279271b4808",
    "CONTAINER_PARTIAL_ORDINAL": "2",
    "CONTAINER_ID_FULL": "fd973ef56c295e4cbfcee89144320326d4185ab5ddffbf4f8bd3cbb08b77f441",
    "__CURSOR": "s=52d6bbae8b5445f7a4155820202b5da6;i=7822da;b=03ff567220e2431795e42279271b4808;m=42aa4aa69;t=620589458c910;x=6ded1baa6a60f040",
    "CONTAINER_PARTIAL_ID": "0c4de2f5604ca894fe3896bbc0ecf9477ffbdea5967e53f6223fdeb68364f80c",
    "_SYSTEMD_SLICE": "system.slice",
    "PRIORITY": "6",
    "_SOURCE_REALTIME_TIMESTAMP": "1724414678319497",
    "COM_GITHUB_LOGUNIFIER_APPLICATION_NAME": "amovabi_services.devicetracking_service[0]",
    "COM_HASHICORP_NOMAD_JOB_ID": "amovabi_services",
    "__REALTIME_TIMESTAMP": "1724414678321424",
    "COM_HASHICORP_NOMAD_NODE_NAME": "worker-01",
    "_UID": "0",
    "IMAGE_NAME": "registry.cloud.private/amovabi/device-tracking-service:2.0.0-CR23",
    "COM_HASHICORP_NOMAD_TASK_GROUP_NAME": "devicetracking_service",
    "__MONOTONIC_TIMESTAMP": "17895303785",
    "MESSAGE": "'1c2375cc-9e7e-4efa-96f1-72c97e65de57'::uuid), ('2024-06-30 16:23:49.389+00'), ('2024-06-30 16:23:46.975+00'), ('2024-06-30 16:23:46.975+00'), ('2024-06-30 16:23:48.299+00'), ('2024-06-30 16:23:48.299+00'), ('CLC227'), (NULL), ('TP227'), ('CDT227'), (NULL), (NULL), ('AUTO'), ('AUTO'), ('OK'), ('OK'), ('AUTO'), ('AUTO'), ('1.324'::double precision), ('0.0'::double precision), ('1.324'::double precision), ('NOT_SET'), ('NOT_SET'), ('0.0,0.0,8500.0'), ('0.0,0.0,2270.0'), ('-1,-1,-1'), ('-1,-1,-1'), ('TRUE'::boolean), ('FALSE'::boolean), ('FALSE'::boolean), ('TRUE'::boolean), ('224052528\\u0000'), (NULL), ('0'::int8), ('0'::int8), ('0'::int8), ('{\\\"moveDirection\\\":\\\"ETL_INPUT->ETL\\\",\\\"PICK_FROM\\\":\\\"ETL_INPUT\\\",\\\"PLACE_PLACE_TO\\\":\\\"ETL\\\"}'), ('{\\\"WEIGHT_CLASS\\\":\\\"7000\\\",\\\"WEIGHT_KG\\\":\\\"7447.0\\\"}')\\n) was aborted: ERROR: invalid byte sequence for encoding \\\"UTF8\\\": 0x00\\n  Where: unnamed portal parameter $32  Call getNextException to see other errors in the batch.\",\"type\":\"org.springframework.dao.DataIntegrityViolationException\"}}",
    "COM_HASHICORP_NOMAD_NAMESPACE": "default",
    "_HOSTNAME": "worker-01",
    "COM_GITHUB_LOGUNIFIER_APPLICATION_PATTERN_KEY": "ecs",
    "_PID": "1251",
    "_CAP_EFFECTIVE": "1ffffffffff",
    "_SELINUX_CONTEXT": "unconfined\n",
    "COM_HASHICORP_NOMAD_JOB_NAME": "amovabi_services",
    "COM_HASHICORP_NOMAD_ALLOC_ID": "3607011c-319e-9e2b-db2a-3ed3bc7f1f79",
    "_EXE": "/usr/bin/dockerd",
    "SYSLOG_TIMESTAMP": "2024-08-23T12:04:38.319435796Z",
    "CONTAINER_ID": "fd973ef56c29",
    "COM_GITHUB_LOGUNIFIER_APPLICATION_ENV": "dev"
  }
`
	converter := JournaldDToEcsConverter{}

	metaPart1 := converter.ConvertToMetaLog(&nats.Msg{
		Subject: "test",
		Header:  nil,
		Data:    []byte(part1),
		Sub:     nil,
	})

	embeddedJsonMessage := strings.ReplaceAll(testNatviceEcs, "\"", "\\\"")
	noPart := []byte(strings.ReplaceAll(testJournaldEcs, "##MSG##", embeddedJsonMessage))
	metaNoPart := converter.ConvertToMetaLog(&nats.Msg{
		Subject: "test",
		Header:  nil,
		Data:    noPart,
		Sub:     nil,
	})
	metaPart2 := converter.ConvertToMetaLog(&nats.Msg{
		Subject: "test",
		Header:  nil,
		Data:    []byte(part2),
		Sub:     nil,
	})
	if metaNoPart.MetaLog.HasProcessErrors() {
		t.Errorf("Expect no Process error But got %s", metaNoPart.MetaLog.EcsLogEntry.ProcessError)
	}

	if !metaPart1.Skip {
		t.Errorf("Expect skip value true. But got %v", metaPart1.Skip)
	}
	if reflect.ValueOf(metaPart2).IsZero() {
		t.Errorf("Expect not nil value. But got %v", metaPart2)
	}
	if metaPart2.Skip {
		t.Errorf("Expect skip value false. But got %v", metaPart2.Skip)
	}

	// metaPart1 +metaPart2 are not valid ECS
	if !metaPart2.MetaLog.HasProcessErrors() {
		t.Errorf("Expect no Process error But got %s", metaPart2.MetaLog.EcsLogEntry.ProcessError)
	}
}
