package model

import (
	"github.com/google/uuid"
	"strings"
)

func UUID() string {
	return uuid.NewString()
}

func StringToLogLevel(level string) LogLevel {
	lowercased := strings.ToLower(level)
	loglevel, found := stringToLogLevelMap[lowercased]
	if !found {
		return stringToLogLevelMap[string(LogLevel_unknown)]
	}
	return loglevel
}

var logLevelToStringMap = map[LogLevel]string{
	LogLevel_trace:   "trace",
	LogLevel_debug:   "debug",
	LogLevel_info:    "info",
	LogLevel_warn:    "warn",
	LogLevel_error:   "error",
	LogLevel_fatal:   "fatal",
	LogLevel_unknown: "unknown",
}

func StringToLogPatterKey(pattern string) MetaLog_PatternKey {
	lowercased := strings.ToLower(pattern)
	key, found := logPatternStringMap[lowercased]
	if !found {
		return logPatternStringMap["nop"]
	}
	return key
}

var logPatternStringMap = map[string]MetaLog_PatternKey{
	"nop":    MetaLog_Nop,
	"logfmt": MetaLog_LogFmt,
	"ecs":    MetaLog_Ecs,
}

var stringToLogLevelMap = map[string]LogLevel{
	// Sync the chaanges here with the log level pattern LOGLEVEL_KEYWORD
	"trace":                  LogLevel_trace,
	"trc":                    LogLevel_trace,
	"debug":                  LogLevel_debug,
	"dbg":                    LogLevel_debug,
	"dbug":                   LogLevel_debug,
	"info":                   LogLevel_info,
	"inf":                    LogLevel_info,
	"notice":                 LogLevel_info,
	"warn":                   LogLevel_warn,
	"warning":                LogLevel_warn,
	"error":                  LogLevel_error,
	"err":                    LogLevel_error,
	"alert":                  LogLevel_error,
	"fatal":                  LogLevel_fatal,
	"emerg":                  LogLevel_fatal,
	"crit":                   LogLevel_fatal,
	"critical":               LogLevel_fatal,
	string(LogLevel_unknown): LogLevel_unknown,
}

var loglevelToEmoji = map[LogLevel]string{
	LogLevel_trace:   "👀",
	LogLevel_debug:   "🐞",
	LogLevel_info:    "✅",
	LogLevel_warn:    "⚠",
	LogLevel_error:   "❌",
	LogLevel_fatal:   "📛",
	LogLevel_unknown: "🤷",
}

func LogLevelToString(level LogLevel) string {
	loglevel, found := logLevelToStringMap[level]
	if !found {
		return logLevelToStringMap[LogLevel_unknown]
	}
	return loglevel
}

func LogLevelToEmoji(level LogLevel) string {
	loglevel, found := loglevelToEmoji[level]
	if !found {
		return loglevelToEmoji[LogLevel_unknown]
	}
	return loglevel
}

func NewEcsLogEntry() *EcsLogEntry {

	return &EcsLogEntry{
		Id:           UUID(),
		Timestamp:    nil,
		Message:      "",
		Tags:         nil,
		Labels:       make(map[string]string),
		Version:      nil,
		Container:    nil,
		Agent:        nil,
		Host:         nil,
		Trace:        nil,
		Organization: nil,
		Service:      nil,
		Error:        nil,
		Log:          &Log{},
		ProcessError: nil,
	}
}
