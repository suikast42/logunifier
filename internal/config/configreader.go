package config

import (
	"flag"
	"fmt"
	"github.com/peterbourgon/ff/v3"
	"github.com/rs/zerolog"
	zeroLogger "github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"os"
	"strings"
	"time"
)

// region read config
func ReadConfigs() {
	fs := flag.NewFlagSet("logunifer", flag.ContinueOnError)

	var (
		natsServers     arrayFlags
		ingressJournalD = fs.String("ingressJournalD", "ingress.logs.journald", "Nats subscription for journald logs")
		egressSubject   = fs.String("egressSubject", "egress.logs.ecs", "Standardized logs output")
		loglevel        = fs.String("loglevel", "info", "Default log level")
		ackTimeoutIns   = fs.Int("ackTimeoutIns", 10, "Ack timeout of ingress channels")
		_               = fs.String("config", "internal/config/local.cfg", "config file (optional)")
	)

	fs.Var(&natsServers, "natsServers", "list of server host and port")
	if err := ff.Parse(fs, os.Args[1:],
		ff.WithEnvVarPrefix("LOGU"),
		ff.WithConfigFileFlag("config"),
		ff.WithConfigFileParser(ff.PlainParser)); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	builder := newBuilder().withIngress(ingressJournalD, IngressLogsJournald)
	for _, s := range natsServers {
		builder.withNatsServer(s)
	}
	_ = builder.
		withLogLevel(loglevel).
		withAckTimeout(ackTimeoutIns).
		withEgressSubject(egressSubject).
		build()

}

type arrayFlags []string

func (i *arrayFlags) String() string {
	return strings.Join(*i, ";")
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

//endregion

//region configure logging

func ConfigLogging() error {
	config, err := Instance()
	if err != nil {
		return err
	}
	// UNIX Time is faster and smaller than most timestamps
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	_ = zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	log := Logger()
	{
		loglevel := config.Loglevel()
		level, err := zerolog.ParseLevel(loglevel)

		if err != nil {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
			log.Warn().Msgf("Can't parse supported loglevel %s use default %s", config.loglevel, zerolog.DebugLevel)
			//return err
		}
		zerolog.SetGlobalLevel(level)
	}
	//log.Trace().Msg("Trace")
	//log.Debug().Msg("Debug")
	//log.Warn().Msg("Warn")
	//log.Info().Msg("Info")
	//log.Error().Msg("Error")
	//log.Fatal().Msg("Fatal")

	log.Info().Msgf("Use configuration:\n%v", config)
	return nil
}

func Logger() zerolog.Logger {
	//TODO provide the ability to switch log output in production to json file
	consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout}

	//multi := zerolog.MultiLevelWriter(consoleWriter, os.Stdout)
	//
	//logger := zerolog.New(multi).With().Timestamp().Logger()

	return zeroLogger.Output(consoleWriter)

}

//endregion
