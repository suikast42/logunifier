package utils

import (
	"bytes"
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
const LogfmtKeyTrash LogFmtKey = "trash"

// DecodeLogFmt makes full usage of lokis logfmt package for regular key value demerited log texts
// All irregular parts of the log are captured in mapkey LogfmtKeyTrash. But if the log does not contain
// any key LogfmtKeyMessage then the log is present in that key and a parse error will return
// Scan all kv pairs
func DecodeLogFmt(log string) (map[string]string, error) {
	var restpart = log
	var parseError error
	result := make(map[string]string)
	if len(log) == 0 {
		parseError = errors.New("empty log not expected")
		return result, parseError
	}
	decoder := logfmt.NewDecoder([]byte(log))

	var trashBuffer bytes.Buffer
	var trashCaught = false
	for decoder.ScanKeyval() {
		if decoder.Err() != nil {
			parseError = errors.Join(parseError, decoder.Err())
			continue
		}
		if decoder.Key() == nil {
			parseError = errors.Join(parseError, errors.New("caught nil key"))
			continue
		}

		newRestpart, isKey := isKey(string(decoder.Key()), restpart)
		restpart = newRestpart
		if decoder.Value() == nil && !isKey {
			if !trashCaught {
				trashCaught = true
			} else {
				trashBuffer.WriteString(" ")
			}
			trashBuffer.WriteString(normalizeKeys(decoder.Key()))

		} else {
			if currentValue, ok := result[normalizeKeys(decoder.Key())]; ok {
				result[normalizeKeys(decoder.Key())] = currentValue + " " + string(decoder.Value())
			} else {
				result[normalizeKeys(decoder.Key())] = string(decoder.Value())
			}
		}
	}

	if len(result) == 0 {
		// That log is marked with parse errors and shows the trash as message
		// Copy the trash to message
		parseError = errors.Join(parseError, errors.New("could not extract key value pairs"))
		result[string(LogfmtKeyMessage)] = log
	} else if trashCaught {
		if len(result[string(LogfmtKeyMessage)]) == 0 {
			// Whole message is parsed and there is no message filed captured
			// make the trash as log message and mark a process error
			result[string(LogfmtKeyMessage)] = trashBuffer.String()
			parseError = errors.Join(parseError, errors.New("is not in logfmt"))
		} else {
			result[string(LogfmtKeyTrash)] = trashBuffer.String()
			parseError = errors.Join(parseError, errors.New("log fmt trash caught"))
		}
	}
	return result, parseError
}

func isKey(word string, wholeWord string) (string, bool) {
	defer func() {
		if r := recover(); r != nil {
			// panic can happen when word and wholeWord changed for example
			// ignore this cases
		}
	}()
	fields := strings.Fields(wholeWord)
	fieldsLen := len(fields)

	var isAKey = false
	for index, current := range fields {
		if strings.Contains(current, word) {
			isAKey = strings.Contains(current, "=")
			return strings.Join(fields[index+1:fieldsLen], " "), isAKey
		}
	}

	//index := strings.Index(wholeWord, word)
	//indexWithEq := strings.Index(wholeWord, word+"=")
	//if index >= 0 {
	//	fmt.Printf("start: %d. end: %d.\n\tWord: [%s], Restword:[%s]\n", len(word)+index, len(wholeWord)-1, wholeWord, wholeWord[len(word)+index:len(wholeWord)-1])
	//	return wholeWord[len(word)+index : len(wholeWord)-1], index == indexWithEq
	//}
	return wholeWord, false
}

//func DecodeLogFmt(log string) (map[string]string, error) {
//
//	decoder := logfmt.NewDecoder(strings.NewReader(log))
//	result := make(map[string]string)
//	var parseError error
//	for decoder.ScanRecord() {
//		for decoder.ScanKeyval() {
//			if decoder.Err() != nil {
//				parseError = errors.Join(parseError, decoder.Err())
//				continue
//			}
//			key := normalizeKeys(decoder.Key())
//
//			value := string(decoder.Value())
//			result[key] = value
//		}
//
//	}
//
//	return result, parseError
//}

func normalizeKeys(key []byte) string {

	lowerKey := strings.ToLower(string(key))

	switch lowerKey {
	case "ts", "timestamp", "time", "t":
		return string(LogfmtKeyTimestamp)
	case "msg", "message":
		return string(LogfmtKeyMessage)
	case "level":
		return string(LogfmtKeyLevel)
	case "err", "error":
		return string(LogfmtKeyError)
	case "caller":
		return string(LogfmtKeyCaller)
	case "traceid", "tid":
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
