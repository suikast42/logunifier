package patterns

import (
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/suikast42/logunifier/internal/config"
	"github.com/suikast42/logunifier/pkg/model"
	"github.com/trivago/grok"
	additionalPatterns "github.com/trivago/grok/patterns"
	"strings"
	"sync"
)

type attributeKeys string
type TimeFormat string

const (
	NopPattern GrokPatternKey = "NopPattern"
)

// var appPatterns = map[GrokPatternKey] GrokPattern
const (
	TimeformatCommonUTC           = "2006-01-02 15:04:05.000 MST"
	TimeFormatConsulConnect       = "2006-01-02 15:04:05.000"
	TimeFormatKeyCloak            = "2006-01-02 15:04:05,000"
	TimeFormatUtcCommaSecondAndTs = "2006-01-02 15:04:05,000-0700"
)
const (
	timestamp attributeKeys = "timestamp"
	level     attributeKeys = "level"
)

type internalGrokKey string

//const (
//	common_level   internalGrokKey = "COMMON_LEVEL"
//	common_ts      internalGrokKey = "COMMON_TS"
//	common_utc_ts  internalGrokKey = "COMMON_UTC_TS"
//	common_nano_ts internalGrokKey = "COMMON_NANO_TS"
//)

//var APPLOGS = map[string]string{
//	// Aliases for patterns
//	"LOGLEVEL_KEYWORD":       `((?i)trace|(?i)trc|(?i)debug|(?i)dbg|(?i)dbug|(?i)info|(?i)inf|(?i)notice|(?i)warn|(?i)warning|(?i)error|(?i)err|(?i)alert|(?i)fatal|(?i)emerg|(?i)crit|(?i)critical)`,
//	"COMMON_UTC_TS_PATTERN":  `%{YEAR}-%{MONTHNUM}-%{MONTHDAY} %{HOUR}:%{MINUTE}:%{SECOND}.%{INT} %{WORD:timezone}`,
//	"COMMON_NANO_TS_PATTERN": `%{YEAR}-%{MONTHNUM}-%{MONTHDAY}T%{HOUR}:%{MINUTE}:%{SECOND}.%{INT:microseconds}Z`,
//	// Used as grok patterns
//	string(common_level):   `(.*level=|.?)%{LOGLEVEL_KEYWORD:level}`,
//	string(common_ts):      `%{TIMESTAMP_ISO8601:timestamp}`,
//	string(common_utc_ts):  `%{COMMON_UTC_TS_PATTERN:timestamp}`,
//	string(common_nano_ts): `%{COMMON_NANO_TS_PATTERN:timestamp}`,
//}

//var APPLOGS = map[string]string{
//	//"MULTILINE":                  `((\s)*(.*))*`,
//	//string(MSG_ONLY):             `%{MULTILINE:message}`,
//	string(TS_LEVEL):        `%{TIMESTAMP_ISO8601:timestamp} .?%{LOGLEVEL:level}.?`,
//	string(LOGFMT_TS_LEVEL): `(time|ts|t)=[",']?%{TIMESTAMP_ISO8601:timestamp}[",']?.*level=%{LOGLEVEL:level}`,
//	string(LOGFMT_LEVEL_TS): `level=%{LOGLEVEL:level}.*(time|ts|t)=[",']?%{TIMESTAMP_ISO8601:timestamp}[",']?`,
//	// This pattern captures the full elements of connect logs.
//	//string(CONNECT_LOG):         `\[%{TIMESTAMP_ISO8601:timestamp}\]\[%{INT:thread_id}\]\[%{LOGLEVEL:level}\]\[%{DATA:module}\] \[%{DATA:source_file}:%{INT:line_number}\] \[%{DATA:connection_id}\] %{MULTILINE:message}`,
//	// This pattern captures a lite version of connect logs and ignores the thread_id
//	string(CONNECT_LOG): `\[%{TIMESTAMP_ISO8601:timestamp}\].*\[%{LOGLEVEL:level}\]`,
//}

type PatternFactory struct {
	logger    *zerolog.Logger
	patterns  map[string]string
	compilers map[string]*grok.CompiledGrok
}

var mtx sync.Mutex

var instance *PatternFactory

func Instance() *PatternFactory {
	return instance
}
func Initialize() (*PatternFactory, error) {
	mtx.Lock()
	defer mtx.Unlock()
	if instance != nil {
		return instance, nil
	}

	addPatterns := make(map[string]string)
	compiledPatterns := make(map[string]*grok.CompiledGrok)
	{
		err := add(addPatterns, grok.DefaultPatterns)
		// better defined in additionalPatterns.Grok
		for k := range additionalPatterns.Grok {
			val := grok.DefaultPatterns[k]
			if len(val) > 0 {
				delete(addPatterns, k)
			}
		}
		if err != nil {
			panic(err)
		}
	}
	{
		err := add(addPatterns, additionalPatterns.Grok)
		if err != nil {
			panic(err)
		}
	}

	grokConfig := grok.Config{
		Patterns:            addPatterns,
		SkipDefaultPatterns: true,
	}
	g, err := grok.New(grokConfig)
	if err != nil {
		suberror := errors.New(fmt.Sprintf("Can't create new isnatce\n%s", err.Error()))
		return nil, suberror
	}
	for k, v := range addPatterns {
		compiled, err := g.Compile(v)
		if err != nil {
			suberror := errors.New(fmt.Sprintf("Cannot compile %s with value %s\n%s", k, v, err.Error()))
			return nil, suberror
		}
		compiledPatterns[k] = compiled
	}
	logger := config.Logger()
	instance = &PatternFactory{
		patterns:  addPatterns,
		compilers: compiledPatterns,
		logger:    &logger,
	}

	return instance, nil
}

func add(source map[string]string, new map[string]string) error {

	for k, v := range new {
		//key := prefix + "_" + k
		key := k
		// IF the key is already present and the default is different
		if val, ok := source[key]; ok && !strings.EqualFold(val, v) {
			return errors.New(fmt.Sprintf("%s already exists for value %s and %s should aded", key, v, val))
		}
		source[key] = v
	}
	return nil
}

func (factory *PatternFactory) Parse(log *model.MetaLog) *model.EcsLogEntry {
	extractor := factory.findPatternFor(log)
	return ExractFrom(extractor, log)
}

func (factory *PatternFactory) findPatternFor(log *model.MetaLog) GrokPatternExtractor {
	return &GrokPatternDefault{
		GrokPattern: GrokPattern{
			Name: NopPattern,
		},
	}
}
