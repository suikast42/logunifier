package patterns

import (
	"errors"
	"fmt"
	"github.com/trivago/grok"
	additionalPatterns "github.com/trivago/grok/patterns"
	"strings"
	"sync"
	"time"
)

type ParseResult struct {
	LogLevel    string
	TimeStamp   time.Time
	UsedPattern string
}

func (p ParseResult) String() string {
	return fmt.Sprintf("Timestamp: %s LogLevel: %s", p.TimeStamp, p.LogLevel)
}

type PatternKey string
type attributeKeys string
type TimeFormat string

const (
	TimeFormatConsulConnect = "2006-01-02 15:04:05.000"
)
const (
	timestamp attributeKeys = "timestamp"
	level     attributeKeys = "level"
)
const (
	CONNECT_LOG     PatternKey = "CONNECT_LOG"
	LOGFMT_TS_LEVEL PatternKey = "LOGFMT_TS_LEVEL"
	LOGFMT_LEVEL_TS PatternKey = "LOGFMT_LEVEL_TS"
	TS_LEVEL        PatternKey = "TS_LEVEL"
)

var tsFormatMap = map[PatternKey]string{
	TS_LEVEL:        time.RFC3339,
	LOGFMT_TS_LEVEL: time.RFC3339,
	LOGFMT_LEVEL_TS: time.RFC3339,
	CONNECT_LOG:     TimeFormatConsulConnect,
}

var APPLOGS = map[string]string{
	//"MULTILINE":                  `((\s)*(.*))*`,
	//string(MSG_ONLY):             `%{MULTILINE:message}`,
	string(TS_LEVEL):        `%{TIMESTAMP_ISO8601:timestamp} .?%{LOGLEVEL:level}.?`,
	string(LOGFMT_TS_LEVEL): `(time|ts|t)=[",']?%{TIMESTAMP_ISO8601:timestamp}[",']?.*level=%{LOGLEVEL:level}`,
	string(LOGFMT_LEVEL_TS): `level=%{LOGLEVEL:level}.*(time|ts|t)=[",']?%{TIMESTAMP_ISO8601:timestamp}[",']?`,
	// This pattern captures the full elements of connect logs.
	//string(CONNECT_LOG):         `\[%{TIMESTAMP_ISO8601:timestamp}\]\[%{INT:thread_id}\]\[%{LOGLEVEL:level}\]\[%{DATA:module}\] \[%{DATA:source_file}:%{INT:line_number}\] \[%{DATA:connection_id}\] %{MULTILINE:message}`,
	// This pattern captures a lite version of connect logs and ignores the thread_id
	string(CONNECT_LOG): `\[%{TIMESTAMP_ISO8601:timestamp}\].*\[%{LOGLEVEL:level}\]`,
}

type PatternFactory struct {
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

	// end::tagname[]
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
	{
		err := add(addPatterns, APPLOGS)
		if err != nil {
			panic(err)
		}
	}
	//{
	//	err := add(addPatterns, additionalPatterns.Java)
	//	if err != nil {
	//		panic(err)
	//	}
	//}
	//{
	//	err := add(addPatterns, additionalPatterns.LinuxSyslog)
	//	if err != nil {
	//		panic(err)
	//	}
	//}
	//{
	//	err := add(addPatterns, additionalPatterns.PostgreSQL)
	//	if err != nil {
	//		panic(err)
	//	}
	//}
	//{
	//	err := add(addPatterns, additionalPatterns.Rails)
	//	if err != nil {
	//		panic(err)
	//	}
	//}
	//{
	//	err := add(addPatterns, additionalPatterns.Redis)
	//	if err != nil {
	//		panic(err)
	//	}
	//}
	//{
	//	err := add(addPatterns, additionalPatterns.Ruby)
	//	if err != nil {
	//		panic(err)
	//	}
	//}
	//{
	//	err := add(addPatterns, NGNIX)
	//	if err != nil {
	//		panic(err)
	//	}
	//}
	config := grok.Config{
		Patterns:            addPatterns,
		SkipDefaultPatterns: true,
	}
	g, err := grok.New(config)
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

	instance = &PatternFactory{
		patterns:  addPatterns,
		compilers: compiledPatterns,
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

func (factory *PatternFactory) Parse(patternkey PatternKey, text string) (ParseResult, error) {
	compiledGrok := factory.compilers[string(patternkey)]
	if compiledGrok == nil {
		return ParseResult{
			UsedPattern: string(patternkey),
		}, errors.New(fmt.Sprintf("No compiler found for key %s", patternkey))
	}
	parsed := compiledGrok.ParseString(text)
	tsFormat, found := tsFormatMap[patternkey]
	var ts time.Time
	if found {
		tsInMsg, tsinmsg := parsed[string(timestamp)]
		if tsinmsg {
			ts, _ = time.Parse(tsFormat, tsInMsg)
		}
	}
	return ParseResult{
		LogLevel:    parsed[string(level)],
		TimeStamp:   ts,
		UsedPattern: string(patternkey),
	}, nil

}

func (factory *PatternFactory) ParseWitDefaults(defaults ParseResult, patternkey PatternKey, text string) (ParseResult, error) {
	parsed, err := factory.Parse(patternkey, text)
	if err != nil {
		return ParseResult{}, err
	}
	if len(parsed.LogLevel) == 0 {
		parsed.LogLevel = defaults.LogLevel
	}

	if parsed.TimeStamp.IsZero() {
		parsed.TimeStamp = defaults.TimeStamp
	}
	return parsed, nil
}
