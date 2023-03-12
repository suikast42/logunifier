package utils

import (
	"errors"
	"github.com/grafana/loki/pkg/logql/log/logfmt"
	"strings"
)

type LogFmtKey string

const LogfmtKeyTimestamp LogFmtKey = "ts"
const LogfmtKeyLevel LogFmtKey = "level"
const LogfmtKeyMessage LogFmtKey = "msg"
const LogfmtKeyCaller LogFmtKey = "caller"
const LogfmtKeyTraceID LogFmtKey = "traceID"
const LogfmtKeySpanID LogFmtKey = "spanID"
const LogfmtKeyError LogFmtKey = "error"
const LogfmtKeyUser LogFmtKey = "user"
const LogfmtKeyEvent LogFmtKey = "event"

func DecodeLogFmt(log string) (map[string]string, error) {

	decoder := logfmt.NewDecoder([]byte(log))
	result := make(map[string]string)
	var parseError error
	for decoder.ScanKeyval() {
		if decoder.Err() != nil {
			parseError = errors.Join(parseError, decoder.Err())
			continue
		}
		result[normalizeKeys(decoder.Key())] = string(decoder.Value())
	}
	return result, parseError
}

func normalizeKeys(key []byte) string {

	lowerKey := strings.ToLower(string(key))

	switch lowerKey {
	case "ts", "timestamp", "time":
		return string(LogfmtKeyTimestamp)
	case "msg", "message":
		return string(LogfmtKeyMessage)
	case "level":
		return string(LogfmtKeyLevel)
	case "err", "error":
		return string(LogfmtKeyError)
	case "caller":
		return string(LogfmtKeyCaller)
	case "traceid":
		return string(LogfmtKeyTraceID)
	case "spanid":
		return string(LogfmtKeySpanID)
	case "user", "usr":
		return string(LogfmtKeyUser)
	case "event":
		return string(LogfmtKeyEvent)
	default:
		return string(key)
	}
}
