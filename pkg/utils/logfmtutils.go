package utils

import (
	"errors"
	"github.com/grafana/loki/pkg/logql/log/logfmt"
)

func DecodeLogFmt(log string) (map[string]string, error) {

	decoder := logfmt.NewDecoder([]byte(log))
	result := make(map[string]string)
	var parseError error
	for decoder.ScanKeyval() {
		if decoder.Err() != nil {
			parseError = errors.Join(parseError, decoder.Err())
			continue
		}
		result[string(decoder.Key())] = string(decoder.Value())
	}
	return result, parseError
}
