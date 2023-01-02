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
	"time"
)

type ParseResult struct {
	LogLevel    model.LogLevel
	TimeStamp   time.Time
	UsedPattern string
}

func (p ParseResult) String() string {
	return fmt.Sprintf("Timestamp: %s LogLevel: %s", p.TimeStamp, p.LogLevel)
}

type PatternKey struct {
	LogLevelPattern string
	TsPattern       string
	Tsformat        string
	Name            string
}

func (p PatternKey) String() string {
	return fmt.Sprintf("%s with time pattern %s formatted by layout %s", p.LogLevelPattern, p.TsPattern, p.Tsformat)
}

type attributeKeys string
type TimeFormat string

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

const (
	common_level   internalGrokKey = "COMMON_LEVEL"
	common_ts      internalGrokKey = "COMMON_TS"
	common_utc_ts  internalGrokKey = "COMMON_UTC_TS"
	common_nano_ts internalGrokKey = "COMMON_NANO_TS"
)

var APPLOGS = map[string]string{
	// Aliases for patterns
	"LOGLEVEL_KEYWORD":       `((?i)trace|(?i)trc|(?i)debug|(?i)dbg|(?i)dbug|(?i)info|(?i)inf|(?i)notice|(?i)warn|(?i)warning|(?i)error|(?i)err|(?i)alert|(?i)fatal|(?i)emerg|(?i)crit|(?i)critical)`,
	"COMMON_UTC_TS_PATTERN":  `%{YEAR}-%{MONTHNUM}-%{MONTHDAY} %{HOUR}:%{MINUTE}:%{SECOND}.%{INT} %{WORD:timezone}`,
	"COMMON_NANO_TS_PATTERN": `%{YEAR}-%{MONTHNUM}-%{MONTHDAY}T%{HOUR}:%{MINUTE}:%{SECOND}.%{INT:microseconds}Z`,
	// Used as grok patterns
	string(common_level):   `(.*level=|.?)%{LOGLEVEL_KEYWORD:level}`,
	string(common_ts):      `%{TIMESTAMP_ISO8601:timestamp}`,
	string(common_utc_ts):  `%{COMMON_UTC_TS_PATTERN:timestamp}`,
	string(common_nano_ts): `%{COMMON_NANO_TS_PATTERN:timestamp}`,
}
var (
	// NopPattern somethmes logs does not contain either a log level or a ts information use this for ignore this one
	NopPattern = PatternKey{
		Name: "IgnorePattern",
	}
	CommonPattern = PatternKey{
		LogLevelPattern: string(common_level),
		TsPattern:       string(common_ts),
		Tsformat:        time.RFC3339,
		Name:            "CommonPattern",
	}
	//CommonPatternNano = PatternKey{
	//	LogLevelPattern: string(common_level),
	//	TsPattern:       string(common_nano_ts),
	//	Tsformat:        time.RFC3339Nano,
	//	Name:            "CommonPatternNano",
	//}
	CommonUtcPattern = PatternKey{
		LogLevelPattern: string(common_level),
		TsPattern:       string(common_utc_ts),
		Tsformat:        TimeformatCommonUTC,
		Name:            "CommonUtcPattern",
	}
	CommonUtcPatternWithCommaTsAndTz = PatternKey{
		LogLevelPattern: string(common_level),
		TsPattern:       string(common_ts),
		Tsformat:        TimeFormatUtcCommaSecondAndTs,
		Name:            "CommonUtcPatternWithCommaTsAndTz",
	}
	ConsulConnectPattern = PatternKey{
		LogLevelPattern: string(common_level),
		TsPattern:       string(common_ts),
		Tsformat:        TimeFormatConsulConnect,
		Name:            "ConsulConnectPattern",
	}
	KeyCloakPattern = PatternKey{
		LogLevelPattern: string(common_level),
		TsPattern:       string(common_ts),
		Tsformat:        TimeFormatKeyCloak,
		Name:            "KeyCloakPattern",
	}
)

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

func (factory *PatternFactory) Parse(foundFor string, patternkey PatternKey, text string) (ParseResult, error) {
	if patternkey == NopPattern {
		return ParseResult{}, nil
	}
	//compiledGrok := factory.compilers[string(patternkey)]
	//if compiledGrok == nil {
	//	return ParseResult{
	//		UsedPattern: string(patternkey),
	//	}, errors.New(fmt.Sprintf("No compiler found for key %s", patternkey))
	//}
	//parsed := compiledGrok.ParseString(text)
	//tsFormat, found := tsFormatMap[patternkey]
	//var ts time.Time
	//if found {
	//	tsInMsg, tsinmsg := parsed[string(timestamp)]
	//	if tsinmsg {
	//		ts, _ = time.Parse(tsFormat, tsInMsg)
	//	}
	//}
	var wg sync.WaitGroup

	var parsedLevel model.LogLevel
	var parsedTs time.Time
	wg.Add(2)
	// extract timestamp
	go func(patternkey *PatternKey, text *string, wg *sync.WaitGroup) {
		defer wg.Done()
		tsCompiler := factory.compilers[patternkey.TsPattern]
		if tsCompiler == nil {
			factory.logger.Error().Msgf("No ts compiler found for %s", patternkey.String())
			return
		}
		parsed := tsCompiler.ParseString(*text)
		tsInMsg, tsFound := parsed[string(timestamp)]
		if tsFound {
			_ts, err := time.Parse(patternkey.Tsformat, tsInMsg)
			if err != nil {
				factory.logger.Error().Err(err).Msgf("Can't parse timestamp for %s .Text is [%s] and found ts is [%s] with detected pattern %s", foundFor, *text, tsInMsg, patternkey.String())
				return
			}
			parsedTs = _ts
		} else {
			factory.logger.Error().Msgf("Could not found a ts for %s .Text is [%s] with detected pattern %s", foundFor, *text, patternkey.String())

		}
	}(&patternkey, &text, &wg)

	// extract loglevel
	go func(patternkey *PatternKey, text *string, wg *sync.WaitGroup) {
		defer wg.Done()
		logLevelCompiler := factory.compilers[patternkey.LogLevelPattern]
		if logLevelCompiler == nil {
			factory.logger.Error().Msgf("No log level compiler found for %s", patternkey.String())
			return
		}
		parsed := logLevelCompiler.ParseString(*text)
		logLevelInMsg, logLevelFound := parsed[string(level)]
		if !logLevelFound || len(logLevelInMsg) == 0 {
			factory.logger.Error().Msgf("Can't find loglevel for %s .Text is [%s] with detected pattern %s", foundFor, *text, patternkey.String())
			return
		}
		parsedLevel = model.StringToLogLevel(logLevelInMsg)
	}(&patternkey, &text, &wg)
	wg.Wait()
	return ParseResult{
		LogLevel:    parsedLevel,
		TimeStamp:   parsedTs,
		UsedPattern: patternkey.Name,
	}, nil

}

func (factory *PatternFactory) ParseWitDefaults(foundFor string, defaults ParseResult, patternkey PatternKey, text string) (ParseResult, error) {
	if patternkey == NopPattern {
		return defaults, nil
	}
	parsed, err := factory.Parse(foundFor, patternkey, text)
	if err != nil {
		return ParseResult{}, err
	}
	if parsed.LogLevel == model.LogLevel_unknown {
		parsed.LogLevel = defaults.LogLevel
	}

	if parsed.TimeStamp.IsZero() {
		parsed.TimeStamp = defaults.TimeStamp
	}
	return parsed, nil
}
