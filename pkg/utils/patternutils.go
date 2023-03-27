package utils

import (
	"fmt"
	"github.com/suikast42/logunifier/pkg/model"
	"github.com/trivago/grok"
	"time"
)

// region pattern parsing

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

//	var APPLOGS = map[string]string{
//		//"MULTILINE":                  `((\s)*(.*))*`,
//		//string(MSG_ONLY):             `%{MULTILINE:message}`,
//		string(TS_LEVEL):        `%{TIMESTAMP_ISO8601:timestamp} .?%{LOGLEVEL:level}.?`,
//		string(LOGFMT_TS_LEVEL): `(time|ts|t)=[",']?%{TIMESTAMP_ISO8601:timestamp}[",']?.*level=%{LOGLEVEL:level}`,
//		string(LOGFMT_LEVEL_TS): `level=%{LOGLEVEL:level}.*(time|ts|t)=[",']?%{TIMESTAMP_ISO8601:timestamp}[",']?`,
//		// This pattern captures the full elements of connect logs.
//		//string(CONNECT_LOG):         `\[%{TIMESTAMP_ISO8601:timestamp}\]\[%{INT:thread_id}\]\[%{LOGLEVEL:level}\]\[%{DATA:module}\] \[%{DATA:source_file}:%{INT:line_number}\] \[%{DATA:connection_id}\] %{MULTILINE:message}`,
//		// This pattern captures a lite version of connect logs and ignores the thread_id
//		string(CONNECT_LOG): `\[%{TIMESTAMP_ISO8601:timestamp}\].*\[%{LOGLEVEL:level}\]`,
//	}

//%{TIMESTAMP_ISO8601:TimeStamp} (%{LOGLEVEL:Level} %{BRACKETED:Thread})|(%{BRACKETED:Thread) %{LOGLEVEL:Level})

type PatterMatch string

const (
	TimeStamp PatterMatch = "timestamp"
	Level     PatterMatch = "level"
	Message   PatterMatch = "message"
)

var patternMatchKeys = map[string]PatterMatch{
	"timestamp": TimeStamp,
	"level":     Level,
	"message":   Message,
}

const (
	timeFormatIso8001    = "timeFormatIso8001"
	timeFormatYYYY_SLASH = "timeFormatYYYY_SLASH"
	timeFormatApacheLog  = "timeFormatApacheLog"
)

var CustomPatterns = map[string]string{
	"MULTILINE":        `((\s)*(.*))*`,
	"LOGLEVEL_KEYWORD": `((?i)trace|(?i)trc|(?i)debug|(?i)dbg|(?i)dbug|(?i)info|(?i)inf|(?i)notice|(?i)warn|(?i)warning|(?i)error|(?i)err|(?i)alert|(?i)fatal|(?i)ftl|(?i)emerg|(?i)crit|(?i)critical)`,
	"TS_YYMMDD_SLASH":  `%{YEAR}/%{MONTHNUM}/%{MONTHDAY} %{TIME}.%{INT:milliseconds}`,
	"TS_APACHE_LOG":    `%{MONTHDAY}/%{MONTH}/%{YEAR}:%{HOUR}:%{MINUTE}:%{SECOND} ?%{ISO8601_TIMEZONE}`,

	"TS": fmt.Sprintf(""+
		"%%{TIMESTAMP_ISO8601:%s}"+
		"|%%{TS_YYMMDD_SLASH:%s}"+
		"|%%{TS_APACHE_LOG:%s}",
		timeFormatIso8001,
		timeFormatYYYY_SLASH,
		timeFormatApacheLog,
	),
	"GENERIC_TS":                      "%{TS:timestamp}",
	model.MetaLog_TsLevelMsg.String(): `[",',\[]?%{GENERIC_TS}[",',\]]? [",',\[]?%{LOGLEVEL_KEYWORD:level}[",',\]]? %{MULTILINE:message}`,
}

func ParseAndGetRegisteredKey(compiler *grok.CompiledGrok, log string) (map[PatterMatch]string, error) {
	result := make(map[PatterMatch]string)

	parsed := compiler.ParseString(log)

	for k, v := range parsed {
		if _, ok := patternMatchKeys[k]; ok {
			result[patternMatchKeys[k]] = v
			continue
		}

	}

	return result, nil
}

//endregion

// region generic ts parsing
var StandardTimeFormats = []string{
	time.RFC3339Nano,
	time.RFC3339,
	time.UnixDate,
	"2006/01/02 15:04:05.000000",
	"2006-01-02 15:04:05,999-0700",
	"2006-01-02 15:04:05,999",
	time.ANSIC,
	time.RubyDate,
	time.StampMilli,
	time.StampMicro,
	time.StampNano,
	"02/Jan/2006:15:04:05 -0700",
	"02/Jan/2006:15:04:05-0700",
}

var tsFormatCahce = make(map[string]string)

func cachedLayoutForLog(log *model.MetaLog) (string, bool) {
	cahceKey := log.ApplicationName + "@" + log.ApplicationVersion
	ts, found := tsFormatCahce[cahceKey]
	return ts, found
}

func cacheLayoutForLog(log *model.MetaLog, ts string) {
	cahceKey := log.ApplicationName + "@" + log.ApplicationVersion
	tsFormatCahce[cahceKey] = ts
}

func deleteCachedLayoutForLog(log *model.MetaLog) {
	cahceKey := log.ApplicationName + "@" + log.ApplicationVersion
	delete(tsFormatCahce, cahceKey)
}

// ParseTimeUncached with all standardTimeFormats and return the first match
// without a parser error
func ParseTimeUncached(timeString string) (time.Time, string) {
	for _, layout := range StandardTimeFormats {
		parse, err := time.Parse(layout, timeString)
		if err != nil || parse.IsZero() {
			continue
		}
		return parse, layout
	}
	return time.Time{}, ""
}

func ParseTime(log *model.MetaLog, timeString string) time.Time {
	if layout, found := cachedLayoutForLog(log); found {
		// Key is cached
		parse, err := time.Parse(layout, timeString)
		if err != nil || parse.IsZero() {
			// expect that a chanced layout always parses a valid timestamp
			// If not delete it from cache and retry it again
			deleteCachedLayoutForLog(log)
			return ParseTime(log, timeString)
		}
		return parse
	}
	// Key is not cached
	parsed, layout := ParseTimeUncached(timeString)
	if !parsed.IsZero() {
		cacheLayoutForLog(log, layout)
	}

	return parsed
}

//endregion
