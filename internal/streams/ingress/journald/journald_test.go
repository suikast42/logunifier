package journald

import (
	"github.com/rs/zerolog"
	"github.com/suikast42/logunifier/internal/config"
	"github.com/suikast42/logunifier/pkg/model"
	"github.com/suikast42/logunifier/pkg/patterns"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"os"
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

	grok, ok := parsed.Labels[log.EcsLogEntry.Log.PatternKey]

	if !ok {
		t.Error("Can't find pattern")
	}

	if grok != model.MetaLog_LogFmt.String() {
		t.Errorf("Expected pattern %s but got %s", model.MetaLog_LogFmt.String(), grok)
	}
	if parsed.Log == nil {
		t.Error("Expected not nil but got nil")
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

	grok, ok := parsed.Labels[log.EcsLogEntry.Log.PatternKey]

	if !ok {
		t.Error("Can't find pattern")
	}

	if grok != model.MetaLog_LogFmt.String() {
		t.Errorf("Expected pattern [%s] but got [%s]", model.MetaLog_LogFmt.String(), grok)
	}
	if parsed.Log == nil {
		t.Error("Expected not nil but got nil")
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

	grok, ok := parsed.Labels[log.EcsLogEntry.Log.PatternKey]

	if !ok {
		t.Error("Can't find pattern")
	}

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

	grok, ok := parsed.Labels[log.EcsLogEntry.Log.PatternKey]

	if !ok {
		t.Error("Can't find pattern")
	}

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

	grok, ok := parsed.Labels[log.EcsLogEntry.Log.PatternKey]

	if !ok {
		t.Error("Can't find pattern")
	}

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

	grok, ok := parsed.Labels[log.EcsLogEntry.Log.PatternKey]

	if !ok {
		t.Error("Can't find pattern")
	}

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

	grok, ok := parsed.Labels[log.EcsLogEntry.Log.PatternKey]

	if !ok {
		t.Error("Can't find pattern")
	}

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

func TestEcsOverJournald(t *testing.T) {
	log := TestMetaLogFromJournalDFromConst([]byte(testJournaldEcs), t)

	if log.HasProcessErrors() {
		t.Errorf("Metalog with process errors %s", log.EcsLogEntry.ProcessError.String())
	}
	parsedEcs := patternfactory.Parse(log)
	if parsedEcs == nil {
		t.Error("Expected not nil but got nil")
	}

	grok, ok := parsedEcs.Labels[(log.EcsLogEntry.Log.PatternKey)]

	if !ok {
		t.Error("Can't find pattern")
	}

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
