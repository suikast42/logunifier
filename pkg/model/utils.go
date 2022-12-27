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
			Level: "Error",
		},
		ParseError: &ParseError{
			Reason:        ParseError_Unmarshal,
			RawData:       string(msg.Data),
			Subject:       msg.Subject,
			MessageHeader: m,
		},
	}
}
