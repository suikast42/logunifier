package clients

import (
	"encoding/json"
	"os"
	"time"

	"github.com/go-kit/log"
	"github.com/rs/zerolog"
	zeroLogger "github.com/rs/zerolog/log"
	"github.com/suikast42/logunifier/pkg/clients/lokiclient"
)

// zeroLogAdapterLogger zerolog adapter for loki client
func zeroLogAdapterLogger() zerolog.Logger {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMicro
	adapter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339Nano,
		FormatPrepare: func(i map[string]interface{}) error {
			message := i["message"]
			json.Unmarshal([]byte(message.(string)), &i)
			delete(i, "message")
			//i["level"] = "info"
			return nil
		},
	}
	//adapter := logdapter.NewLogfmtWriter(os.Stdout)
	//multi := zerolog.MultiLevelWriter(consoleWriter, os.Stdout)
	//
	//logger := zerolog.New(multi).With().Timestamp().Logger()
	return zeroLogger.Output(adapter)
}

func NewLokiClient(cfg lokiclient.Config) (lokiclient.Client, error) {
	//metrics := lokiclient.NewMetrics(prometheus.DefaultRegisterer)
	metrics := lokiclient.NewMetrics(nil)
	lokiClient, err := lokiclient.New(metrics, cfg, 0, 0, true, log.NewJSONLogger(zeroLogAdapterLogger()))
	if err != nil {
		return nil, err
	}
	return lokiClient, nil
}
