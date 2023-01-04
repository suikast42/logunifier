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
		natsServers            arrayFlags
		lokiServers            arrayFlags
		ingressSubjectJournalD = fs.String("ingressSubjectJournalD", "ingress.logs.journald", "ingress subject journald logs shipped by vector")
		//ingressSubjectDocker   = fs.String("ingressSubjectDocker", "ingress.logs.docker", "ingress subject docker container logs shipped by vector")
		ingressSubjectTest = fs.String("ingressSubjectTest", "ingress.logs.test", "Nats subscription for test logs")
		egressSubjectEcs   = fs.String("egressSubjectEcs", "egress.logs.ecs", "Standardized logs output")
		loglevel           = fs.String("loglevel", "info", "Default log level")
		ackTimeoutIns      = fs.Int("ackTimeoutIns", 10, "Ack timeout of ingress channels")
		_                  = fs.String("config", "internal/config/local.cfg", "config file (optional)")
	)

	// Default defined in local.cfg
	fs.Var(&natsServers, "natsServers", "list of nats server(s) host and port")
	fs.Var(&lokiServers, "lokiServers", "list of loki server(s) host and port")
	if err := ff.Parse(fs, os.Args[1:],
		ff.WithEnvVarPrefix("LOGU"),
		ff.WithConfigFileFlag("config"),
		ff.WithConfigFileParser(ff.PlainParser)); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	builder := newBuilder().
		withIngressSubjectJournald(ingressSubjectJournalD).
		withIngresSubjectTest(ingressSubjectTest)
	//withIngressSubjectDocker(ingressSubjectDocker)
	for _, s := range natsServers {
		builder.withNatsServer(s)
	}
	for _, s := range lokiServers {
		builder.withLokiServers(s)
	}
	_ = builder.
		withLogLevel(loglevel).
		withAckTimeout(ackTimeoutIns).
		withEgressSubjectEcs(egressSubjectEcs).
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
	//_ = zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339Nano}
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

	consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}

	//multi := zerolog.MultiLevelWriter(consoleWriter, os.Stdout)
	//
	//logger := zerolog.New(multi).With().Timestamp().Logger()

	return zeroLogger.Output(consoleWriter)

}

//endregion
