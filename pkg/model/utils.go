package model

import (
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"strings"
)

func UUID() string {
	return uuid.NewString()
}
func ToUnmarshalError(msg *nats.Msg, err error) *EcsLogEntry {
	m := make(map[string]string)
	for k, v := range msg.Header {
		m[k] = strings.Join(v, ",")
	}
	return &EcsLogEntry{
		Message: err.Error(),
		Log: &Log{
			Level: LogLevel_ERROR,
		},
		ParseError: &ParseError{
			Reason:        ParseError_Unmarshal,
			RawData:       string(msg.Data),
			Subject:       msg.Subject,
			MessageHeader: m,
		},
	}
}

func StringToLogLevel(level string) LogLevel {

	lowercased := strings.ToLower(level)

	switch lowercased {
	case "trace", "trc":
		return LogLevel_TRACE
	case "debug", "dbg":
		return LogLevel_DEBUG
	case "info", "inf":
		return LogLevel_INFO
	case "warn", "warning":
		return LogLevel_WARN
	case "error", "alert":
		return LogLevel_ERROR
	case "fatal":
		return LogLevel_FATAL
	default:
		return LogLevel_UNKNOWN
	}
}

func LogLevelToString(level LogLevel) string {

	switch level {
	case LogLevel_TRACE:
		return "TRACE"
	case LogLevel_DEBUG:
		return "DEBUG"
	case LogLevel_INFO:
		return "INFO"
	case LogLevel_WARN:
		return "WARN"
	case LogLevel_ERROR:
		return "ERROR"
	case LogLevel_FATAL:
		return "FATAL"
	case LogLevel_UNKNOWN:
		return "UNKNOWN"
	default:
		return "UNKNOWN"
	}
}

func LogLevelToEmoji(level LogLevel) string {

	switch level {
	case LogLevel_TRACE:
		return "👓"
	case LogLevel_DEBUG:
		return "🐞"
	case LogLevel_INFO:
		return "✔"
	case LogLevel_WARN:
		return "⚠"
	case LogLevel_ERROR:
		return "❌"
	case LogLevel_FATAL:
		return "📛"
	case LogLevel_UNKNOWN:
		return "🤷"
	default:
		return "🤷"
	}
}
