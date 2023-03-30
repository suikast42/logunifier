package patterns

import (
	"github.com/rs/zerolog"
	"github.com/suikast42/logunifier/internal/config"
	"github.com/suikast42/logunifier/pkg/model"
	"github.com/suikast42/logunifier/pkg/utils"
	"os"
	"reflect"
	"testing"
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

func TestGenericTsPattern(t *testing.T) {
	tests := []struct {
		pos        int
		patternKey string
		data       string
	}{
		{
			pos:        1,
			patternKey: "GENERIC_TS",
			data:       "2023-03-20T15:06:45.057Z",
		},
		{
			pos:        2,
			patternKey: "GENERIC_TS",
			data:       "2023-03-20 14:27:28,296",
		},
		{
			pos:        3,
			patternKey: "GENERIC_TS",
			data:       "2023-03-19 21:17:04,243+0000",
		},
		{
			pos:        4,
			patternKey: "GENERIC_TS",
			data:       "2022-08-04T09:53:59.620557561Z",
		},

		{
			pos:        5,
			patternKey: "GENERIC_TS",
			data:       "2023/03/20 14:27:52.652648",
		},

		{
			pos:        6,
			patternKey: "GENERIC_TS",
			data:       "02/Feb/2023:15:04:05 -0700",
		},

		{
			pos:        7,
			patternKey: "GENERIC_TS",
			data:       "2023-03-27T18:23:45Z",
		},
		{
			pos:        8,
			patternKey: "GENERIC_TS",
			data:       "27/Mar/2023:18:23:45-0400",
		},

		{
			pos:        9,
			patternKey: "GENERIC_TS",
			data:       "2023-03-29 20:50:13.931",
		},
	}
	log := &model.MetaLog{
		ApplicationName:    "Test",
		ApplicationVersion: "1",
	}
	for _, test := range tests {
		kv, parseError := patternfactory.ParseGrokWithKey(test.patternKey, test.data)
		if parseError != nil {
			t.Errorf("Pos: %d for pattern %s. Parse errror %s", test.pos, test.patternKey, parseError)
			continue
		}

		parsedTs := kv[utils.PatternMatchTimeStamp]

		if len(parsedTs) != len(test.data) {
			t.Errorf("In pos: %d. Expect [%s] but got [%s]. Parsed %+v", test.pos, test.data, parsedTs, kv)
			continue
		}

		convertedTs := utils.ParseTime(log, parsedTs)

		if convertedTs.IsZero() {
			t.Errorf("Pos: %d. Ts is zero for ts %s", test.pos, parsedTs)
		}
	}
}

func TestPatterns(t *testing.T) {
	tests := []struct {
		pos        int
		patternKey model.MetaLog_PatternKey
		data       string
		want       map[utils.PatterMatch]string
	}{
		{
			pos:        1,
			patternKey: model.MetaLog_TsLevelMsg,
			data:       "2023-03-20T15:06:45.057Z [DEBUG] nomad: memberlist: Stream connection from=127.0.0.1:48046",
			want: map[utils.PatterMatch]string{
				utils.PatternMatchTimeStamp:  "2023-03-20T15:06:45.057Z",
				utils.PatternMatchKeyLevel:   "DEBUG",
				utils.PatternMatchKeyMessage: "nomad: memberlist: Stream connection from=127.0.0.1:48046",
			},
		},

		{
			pos:        2,
			patternKey: model.MetaLog_TsLevelMsg,
			data:       "[2023-03-20T15:06:45.057Z] DEBUG nomad: memberlist: Stream connection from=127.0.0.1:48046",
			want: map[utils.PatterMatch]string{
				utils.PatternMatchTimeStamp:  "2023-03-20T15:06:45.057Z",
				utils.PatternMatchKeyLevel:   "DEBUG",
				utils.PatternMatchKeyMessage: "nomad: memberlist: Stream connection from=127.0.0.1:48046",
			},
		},

		{
			pos:        3,
			patternKey: model.MetaLog_TsLevelMsg,
			data:       "[2023-03-20T15:06:45.057Z] [DEBUG] nomad: memberlist: Stream connection from=127.0.0.1:48046",
			want: map[utils.PatterMatch]string{
				utils.PatternMatchTimeStamp:  "2023-03-20T15:06:45.057Z",
				utils.PatternMatchKeyLevel:   "DEBUG",
				utils.PatternMatchKeyMessage: "nomad: memberlist: Stream connection from=127.0.0.1:48046",
			},
		},

		{
			pos:        4,
			patternKey: model.MetaLog_TsLevelMsg,
			data:       "2023-03-20T15:06:45.057Z [DEBUG] nomad: memberlist: Stream connection from=127.0.0.1:48046",
			want: map[utils.PatterMatch]string{
				utils.PatternMatchTimeStamp:  "2023-03-20T15:06:45.057Z",
				utils.PatternMatchKeyLevel:   "DEBUG",
				utils.PatternMatchKeyMessage: "nomad: memberlist: Stream connection from=127.0.0.1:48046",
			},
		},
		{
			// nomad
			// consul
			pos:        5,
			patternKey: model.MetaLog_TsLevelMsg,
			data:       "2023-03-20T15:06:45.057Z DEBUG nomad: memberlist: Stream connection from=127.0.0.1:48046",
			want: map[utils.PatterMatch]string{
				utils.PatternMatchTimeStamp:  "2023-03-20T15:06:45.057Z",
				utils.PatternMatchKeyLevel:   "DEBUG",
				utils.PatternMatchKeyMessage: "nomad: memberlist: Stream connection from=127.0.0.1:48046",
			},
		},
		{
			// Nexus
			pos:        6,
			patternKey: model.MetaLog_TsLevelMsg,
			data:       "2023-03-19 21:17:04,243+0000 INFO [FelixStartLevel] *SYSTEM ROOT - bundle org.apache.felix.scr:2.1.30 (54) Starting with globalExtender setting: false",
			want: map[utils.PatterMatch]string{
				utils.PatternMatchTimeStamp:  "2023-03-19 21:17:04,243+0000",
				utils.PatternMatchKeyLevel:   "INFO",
				utils.PatternMatchKeyMessage: "[FelixStartLevel] *SYSTEM ROOT - bundle org.apache.felix.scr:2.1.30 (54) Starting with globalExtender setting: false",
			},
		},
		{
			//keycloak
			pos:        7,
			patternKey: model.MetaLog_TsLevelMsg,
			data:       "2023-03-20 14:27:28,296 INFO [org.infinispan.CLUSTER] (keycloak-cache-init) ISPN000079: Channel `ISPN` local address is `b52fd99994da-52866`, physical addresses are `[172.26.68.59:37184]`",
			want: map[utils.PatterMatch]string{
				utils.PatternMatchTimeStamp:  "2023-03-20 14:27:28,296",
				utils.PatternMatchKeyLevel:   "INFO",
				utils.PatternMatchKeyMessage: "[org.infinispan.CLUSTER] (keycloak-cache-init) ISPN000079: Channel `ISPN` local address is `b52fd99994da-52866`, physical addresses are `[172.26.68.59:37184]`",
			},
		},
		{
			//nats
			pos:        8,
			patternKey: model.MetaLog_TsLevelMsg,
			data:       "[1] 2023/03/20 14:27:52.652648 [INF] Server is ready",
			want: map[utils.PatterMatch]string{
				utils.PatternMatchTimeStamp:  "2023/03/20 14:27:52.652648",
				utils.PatternMatchKeyLevel:   "INF",
				utils.PatternMatchKeyMessage: "Server is ready",
			},
		},

		{
			//Apache log ts
			pos:        8,
			patternKey: model.MetaLog_TsLevelMsg,
			data:       "02/Feb/2023:15:04:05 -0700 [INF] Server is ready",
			want: map[utils.PatterMatch]string{
				utils.PatternMatchTimeStamp:  "02/Feb/2023:15:04:05 -0700",
				utils.PatternMatchKeyLevel:   "INF",
				utils.PatternMatchKeyMessage: "Server is ready",
			},
		},
		{
			//W3c log ts
			pos:        8,
			patternKey: model.MetaLog_TsLevelMsg,
			data:       "2023-03-27T18:23:45Z [INF] Server is ready",
			want: map[utils.PatterMatch]string{
				utils.PatternMatchTimeStamp:  "2023-03-27T18:23:45Z",
				utils.PatternMatchKeyLevel:   "INF",
				utils.PatternMatchKeyMessage: "Server is ready",
			},
		},
		{
			pos:        9,
			patternKey: model.MetaLog_TsLevelMsg,
			data:       "2023-03-29 20:50:13.931 [INF] Server is ready",
			want: map[utils.PatterMatch]string{
				utils.PatternMatchTimeStamp:  "2023-03-29 20:50:13.931",
				utils.PatternMatchKeyLevel:   "INF",
				utils.PatternMatchKeyMessage: "Server is ready",
			},
		},
		{
			//Envoy log
			pos:        10,
			patternKey: model.MetaLog_Envoy,
			data:       "[2023-03-30 10:51:43.705][42][debug][upstream] [source/common/upstream/upstream_impl.cc:451] transport socket match, socket default selected for host with address 10.21.21.42:9411",
			want: map[utils.PatterMatch]string{
				utils.PatternMatchTimeStamp:  "2023-03-30 10:51:43.705",
				utils.PatternMatchKeyLevel:   "debug",
				utils.PatternMatchKeyMessage: "[upstream] [source/common/upstream/upstream_impl.cc:451] transport socket match, socket default selected for host with address 10.21.21.42:9411",
				utils.PatternMatchKeyThread:  "42",
			},
		},
		{
			//Logunifier log
			pos:        11,
			patternKey: model.MetaLog_TsLevelMsg,
			data:       "2023-03-30T16:32:12.538785+02:00 INF Connected to Loki. State is READY",
			want: map[utils.PatterMatch]string{
				utils.PatternMatchTimeStamp:  "2023-03-30T16:32:12.538785+02:00",
				utils.PatternMatchKeyLevel:   "INF",
				utils.PatternMatchKeyMessage: "Connected to Loki. State is READY",
			},
		},
	}

	for _, test := range tests {
		kv, err := patternfactory.ParseGrok(test.patternKey, test.data)

		if err != nil {
			t.Errorf("Pos: %d. No error expexted but got %s", test.pos, err)
			continue
		}

		if len(kv) != len(test.want) {
			t.Errorf("In pos: %d. Expect %d keys but got %d", test.pos, len(test.want), len(kv))
		}
		if !reflect.DeepEqual(kv, test.want) {
			t.Errorf("\npos:%d \nin: %q\nwant: %+v\ngot:  %+v", test.pos, test.data, test.want, kv)
		}

		if err != nil {
			t.Errorf("Pos: %d. Can't parse ts [%s]. %s", test.pos, kv[utils.PatternMatchTimeStamp], err)
		}

		parsedTs, _ := utils.ParseTimeUncached(kv[utils.PatternMatchTimeStamp])

		if parsedTs.IsZero() {
			t.Errorf("Pos: %d. Can't parse ts [%s]. parsedTs.IsZero", test.pos, kv[utils.PatternMatchTimeStamp])
		}
	}
}

func TestTimeParseTimeZone(t *testing.T) {
	tests := []struct {
		pos    int
		data   string
		year   int
		month  string
		day    int
		hour   int
		minute int
		second int
	}{
		{
			pos:    1,
			data:   "2023-03-29T20:30:00+0000",
			year:   2023,
			month:  "March",
			day:    29,
			hour:   20,
			minute: 30,
			second: 00,
		},
		{
			pos:    2,
			data:   "2023-03-29T20:30:00+0200",
			year:   2023,
			month:  "March",
			day:    29,
			hour:   18,
			minute: 30,
			second: 00,
		},
		{
			pos:    3,
			data:   "2023-03-29T20:30:00-0200",
			year:   2023,
			month:  "March",
			day:    29,
			hour:   22,
			minute: 30,
			second: 00,
		},
		{
			pos:    4,
			data:   "2023-03-29 20:50:13.931",
			year:   2023,
			month:  "March",
			day:    29,
			hour:   20,
			minute: 50,
			second: 13,
		},
	}
	for _, test := range tests {
		time, layout := utils.ParseTimeUncached(test.data)
		//fmt.Printf("Pos %d: %s\n", test.pos, time.String())
		//continue
		if time.IsZero() {
			t.Errorf("Pos %d: Expected time not zero  but got %+v for layout [%s]", test.pos, time, layout)
		}
		if time.Year() != test.year {
			t.Errorf("Pos %d: Expected Year %d  but got %d", test.pos, test.year, time.Year())
		}
		if time.Month().String() != test.month {
			t.Errorf("Pos %d: Expected Month %s  but got %s", test.pos, test.month, time.Month())
		}
		if time.Day() != test.day {
			t.Errorf("Pos %d: Expected Day %d  but got %d", test.pos, test.day, time.Day())
		}
		if time.Minute() != test.minute {
			t.Errorf("Pos %d: Expected  Minute %d  but got %d", test.pos, test.minute, time.Minute())
		}
		if time.Hour() != test.hour {
			t.Errorf("Pos %d: Expected Hour %d  but got %d", test.pos, test.hour, time.Hour())
		}
		if time.Second() != test.second {
			t.Errorf("Pos %d: Expected Second %d  but got %d", test.pos, test.second, time.Second())
		}
	}
}
