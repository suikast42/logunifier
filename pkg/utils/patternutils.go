package utils

import (
	"fmt"
	"github.com/suikast42/logunifier/pkg/model"
	"github.com/trivago/grok"
	"time"
)

// region pattern parsing

type PatterMatch string

const (
	PatternMatchTimeStamp     PatterMatch = "timestamp"
	PatternMatchKeyLevel      PatterMatch = "level"
	PatternMatchKeyMessage    PatterMatch = "message"
	PatternMatchKeyThread     PatterMatch = "thread"
	PatternMatchKeyOrigin     PatterMatch = "origin"
	PatternMatchKeyOriginLine PatterMatch = "originline"
)

var patternMatchKeys = map[string]PatterMatch{
	"timestamp":  PatternMatchTimeStamp,
	"level":      PatternMatchKeyLevel,
	"message":    PatternMatchKeyMessage,
	"thread":     PatternMatchKeyThread,
	"origin":     PatternMatchKeyOrigin,
	"originline": PatternMatchKeyOriginLine,
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
	model.MetaLog_Envoy.String():      `[",',\[]?%{GENERIC_TS}[",',\]]?[",',\[]?%{NUMBER:thread}[",',\]]?[",',\[]?%{LOGLEVEL_KEYWORD:level}[",',\]]?%{MULTILINE:message}`,
	model.MetaLog_TsLevelMsg.String(): `[",',\[]?%{GENERIC_TS}[",',\]]? [",',\[]?%{LOGLEVEL_KEYWORD:level}[",',\]]? %{MULTILINE:message}`,
	model.MetaLog_Clf.String():        `%{IPORHOST:client_ip} %{USER:ident} %{USER:auth} \[%{HTTPDATE:timestamp}\] \"%{WORD:method} %{URIPATHPARAM:request} HTTP/%{NUMBER:http_version}\" %{NUMBER:status_code} %{NUMBER:bytes} \"%{DATA:referrer}\" \"%{DATA:user_agent}\"`,
	model.MetaLog_Traefik.String():    `%{TIMESTAMP_ISO8601:timestamp} %{LOGLEVEL_KEYWORD:level} %{DATA:origin}:%{NUMBER:originline} > %{GREEDYDATA:message}`,
}

func ParseAndGetRegisteredKey(compiler *grok.CompiledGrok, log string) (map[PatterMatch]string, error) {
	result := make(map[PatterMatch]string)

	parsed := compiler.ParseString(log)

	for k, v := range parsed {
		//result[patternMatchKeys[k]] = v
		if IsRegisteredKey(k) {
			result[patternMatchKeys[k]] = v
		}

	}

	return result, nil
}

func IsRegisteredKey(key string) bool {
	if _, ok := patternMatchKeys[key]; ok {
		return ok
	}
	return false
}

//endregion

// region generic ts parsing
var StandardTimeFormats = []string{
	time.RFC3339Nano,
	time.RFC3339,
	time.UnixDate,
	"2006/01/02 15:04:05.000000",
	"2006-01-02 15:04:05,999-0700",
	"2006-01-02 15:04:05,999 -0700",
	"2006-01-02T15:04:05-0700",
	"2006-01-02T15:04:05 -0700",
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
	ts, found := tsFormatCahce[cacheKeyForLog(log)]
	return ts, found
}

func cacheLayoutForLog(log *model.MetaLog, ts string) {
	tsFormatCahce[cacheKeyForLog(log)] = ts
}

func deleteCachedLayoutForLog(log *model.MetaLog) {
	delete(tsFormatCahce, cacheKeyForLog(log))
}

func cacheKeyForLog(log *model.MetaLog) string {
	if log.EcsLogEntry != nil && log.EcsLogEntry.Service != nil {
		return log.EcsLogEntry.Service.Name + "@" + log.EcsLogEntry.Service.Version
	}
	return "NoService@NoVersion"
}

// ParseTimeUncached with all standardTimeFormats and return the first match
// without a parser error
func ParseTimeUncached(timeString string) (time.Time, string) {
	for _, layout := range StandardTimeFormats {
		//parse, err := time.Parse(layout, timeString)
		parse, err := time.ParseInLocation(layout, timeString, time.UTC)
		if err != nil || parse.IsZero() {
			continue
		}
		return parse.UTC(), layout
	}
	return time.Time{}.UTC(), ""
}

func ParseTime(log *model.MetaLog, timeString string) time.Time {
	if layout, found := cachedLayoutForLog(log); found {
		// Key is cached
		//parse, err := time.Parse(layout, timeString)
		parse, err := time.ParseInLocation(layout, timeString, time.UTC)
		if err != nil || parse.IsZero() {
			// expect that a chanced layout always parses a valid timestamp
			// If not delete it from cache and retry it again
			deleteCachedLayoutForLog(log)
			return ParseTime(log, timeString)
		}
		return parse.UTC()
	}
	// Key is not cached
	parsed, layout := ParseTimeUncached(timeString)
	if !parsed.IsZero() {
		cacheLayoutForLog(log, layout)
	}

	return parsed.UTC()
}

//endregion
