package main

import (
	"github.com/rs/zerolog"
	"github.com/suikast42/logunifier/internal/bootstrap"
	"github.com/suikast42/logunifier/internal/config"
	"github.com/suikast42/logunifier/internal/streams/connectors/lokishipper"
	"github.com/suikast42/logunifier/internal/streams/ingress"
	"github.com/suikast42/logunifier/internal/streams/ingress/dockerlogs"
	"github.com/suikast42/logunifier/internal/streams/ingress/journald"
	"github.com/suikast42/logunifier/internal/streams/ingress/testingress"
	"github.com/suikast42/logunifier/internal/streams/process"
	internalPatterns "github.com/suikast42/logunifier/pkg/patterns"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"
)

func main() {

	//fmt.Printf("GRPC_GO_LOG_VERBOSITY_LEVEL: %s\n", os.Getenv("GRPC_GO_LOG_VERBOSITY_LEVEL")) // 99
	//fmt.Printf("GRPC_GO_LOG_SEVERITY_LEVEL: %s\n", os.Getenv("GRPC_GO_LOG_SEVERITY_LEVEL"))  // info
	processChannel := make(chan ingress.IngressMsgContext, 4096)
	//ctx, cancelFunc := context.WithCancel(context.Background())
	// Listen on os exit signals
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)
	//Read the config file -> program flags
	config.ReadConfigs()
	err := config.ConfigLogging()
	// The first panic point. Must be sure that the config is done
	if err != nil {
		panic(err)
	}

	cfg, err := config.Instance()

	// The second  panic point. Must be sure that the config is done
	if err != nil {
		panic(err)
	}
	logger := config.Logger()
	defer func(logger zerolog.Logger) {
		if r := recover(); r != nil {
			// Log fatal do an os.Exit(1)
			stack := debug.Stack()
			logger.Fatal().Msgf("Unexpected error: %+v\n%s", r, string(stack))
		}
	}(logger)

	logger.Info().Msgf("Starting with config: %s", cfg.String())

	err = process.Start(processChannel, cfg.EgressSubjectEcs())
	if err != nil {
		logger.Error().Err(err).Stack().Msg("Can't start process channel")
		os.Exit(1)
	}
	//Initialize pattern factory
	_, err = internalPatterns.Initialize()
	if err != nil {
		logger.Error().Err(err).Stack().Msg("Can't initialize pattern factory")
		os.Exit(1)
	}

	//Stream definitions
	const (
		streamNameLogStreamIngress = "LogStreamIngress"
		streamNameLogStreamEgress  = "LogStreamEgress"
	)

	streamDefinitions := make(map[string]bootstrap.NatsStreamConfiguration)

	streamDefinitions[streamNameLogStreamIngress] = bootstrap.NatsStreamConfiguration{
		StreamConfiguration: bootstrap.StreamConfig(streamNameLogStreamIngress,
			"Ingress stream for unify and enrich logs from various formats to ecs",
			[]string{
				cfg.IngressNatsJournald(),
				cfg.IngresNatsTest(),
				cfg.IngressNatsDocker(),
			}),
	}
	streamDefinitions[streamNameLogStreamEgress] = bootstrap.NatsStreamConfiguration{
		StreamConfiguration: bootstrap.StreamConfig(streamNameLogStreamEgress,
			"Egress stream that contains ecs logs in json format for ship in various sinks",
			[]string{
				cfg.EgressSubjectEcs(),
			}),
	}

	const (
		ingressConsumerTest     = "ConsumerIngressTest"
		ingressConsumerJournalD = "ConsumerIngressJournalD"
		ingressConsumerDocker   = "ConsumerIngressDocker"
		egressLokiShipper       = "ConsumerEgressLokiShipper"
		egressLokiShipper2      = "ConsumerEgressLokiShipper2"
	)

	// Ingress stream Consumer configuration
	streamConsumerDefinitions := make(map[string]bootstrap.NatsConsumerConfiguration)

	streamConsumerDefinitions[ingressConsumerTest] = bootstrap.NatsConsumerConfiguration{
		ConsumerConfiguration: bootstrap.QueueSubscribeConsumerGroupConfig(
			ingressConsumerTest,
			ingressConsumerTest+"_Group",
			streamDefinitions[streamNameLogStreamIngress].StreamConfiguration,
			cfg.IngresNatsTest(),
		),
		StreamName: streamNameLogStreamIngress,
		MsgHandler: bootstrap.IngressMsgHandler(processChannel, &testingress.TestEcsConverter{}),
	}
	streamConsumerDefinitions[ingressConsumerJournalD] = bootstrap.NatsConsumerConfiguration{
		ConsumerConfiguration: bootstrap.QueueSubscribeConsumerGroupConfig(
			ingressConsumerJournalD,
			ingressConsumerJournalD+"_Group",
			streamDefinitions[streamNameLogStreamIngress].StreamConfiguration,
			cfg.IngressNatsJournald(),
		),
		StreamName: streamNameLogStreamIngress,
		MsgHandler: bootstrap.IngressMsgHandler(processChannel, &journald.JournaldDToEcsConverter{}),
	}
	streamConsumerDefinitions[ingressConsumerDocker] = bootstrap.NatsConsumerConfiguration{
		ConsumerConfiguration: bootstrap.QueueSubscribeConsumerGroupConfig(
			ingressConsumerDocker,
			ingressConsumerDocker+"_Group",
			streamDefinitions[streamNameLogStreamIngress].StreamConfiguration,
			cfg.IngressNatsDocker(),
		),
		StreamName: streamNameLogStreamIngress,
		MsgHandler: bootstrap.IngressMsgHandler(processChannel, &dockerlogs.DockerToEcsConverter{}),
	}
	// Egress stream Consumer configuration
	lokiShipper := &lokishipper.LokiShipper{
		Logger:        config.Logger(),
		LokiAddresses: cfg.LokiServers(),
	}
	lokiShipper.Connect()

	streamConsumerDefinitions[egressLokiShipper] = bootstrap.NatsConsumerConfiguration{
		ConsumerConfiguration: bootstrap.QueueSubscribeConsumerGroupConfig(
			egressLokiShipper,
			egressLokiShipper+"_Group",
			streamDefinitions[streamNameLogStreamEgress].StreamConfiguration,
			cfg.EgressSubjectEcs(),
		),
		StreamName: streamNameLogStreamEgress,
		MsgHandler: bootstrap.EgressMessageHandler(lokiShipper),
	}

	dialer, err := bootstrap.New(streamDefinitions, streamConsumerDefinitions)
	if err != nil {
		logger.Error().Err(err).Msg("Instantiation error during bootstrap")
		os.Exit(1)
	}
	//Connect to nats server
	err = dialer.Connect()
	if err != nil {
		logger.Error().Err(err).Stack().Msg("Can't connect to nats")
		os.Exit(1)
	}

	//processChannel := make(chan *ingress.IngressMsgContext, 4096)

	//TODO find a nicer way to define the JetStream subscriptions
	//subscriptionIngressJournald, err := journald.NewSubscription("IngressLogsJournaldStream", "IngressLogsJournaldProcessor", []string{instance.IngressNatsJournald()}, processChannel)
	//if err != nil {
	//	logger.Error().Err(err).Msgf("Can't subscribe to %v", instance.IngressNatsJournald())
	//	os.Exit(1)
	//}

	//subscriptionIngressTest, err := testingress.NewSubscription("IngressLogsTestStream", "IngressLogsTestStreamProcessor", []string{instance.IngresNatsTest()}, processChannel)
	//if err != nil {
	//	logger.Error().Err(err).Msgf("Can't subscribe to %v", instance.IngresNatsTest())
	//	os.Exit(1)
	//}
	//
	//subscriptionEgressEcsToLoki, err := lokishipper.NewSubscription("EcsToLoki", "EcsToLokiProcessor", []string{instance.EgressSubjectEcs()})
	//if err != nil {
	//	logger.Error().Err(err).Msgf("Can't subscribe to %v", instance.EgressSubjectEcs())
	//	os.Exit(1)
	//}
	//
	////subscriptionEgressEcsToLoki2, err := lokishipper.NewSubscriptionLokiShipper2("EcsToLoki2", "EcsToLokiProcessor2", []string{instance.EgressSubjectEcs()})
	////if err != nil {
	////	logger.Error().Err(err).Msgf("Can't subscribe to %v", instance.EgressSubjectEcs())
	////	os.Exit(1)
	////}
	//
	//subscriptions := []bootstrap.NatsSubscription{
	//	//subscriptionIngressJournald,
	//	subscriptionIngressTest,
	//	subscriptionEgressEcsToLoki,
	//	//subscriptionEgressEcsToLoki2,
	//}
	//
	//// Connect to nats server(s) should be a save background process
	//// It handles reconnection logic itself
	//// If an error occurs here then something is general wrong.
	//// Exit here
	//dialer := bootstrap.New(subscriptions)
	//err = dialer.Connect()
	//if err != nil {
	//	logger.Error().Err(err).Stack().Msg("Can't connect to nats")
	//	os.Exit(1)
	//}
	//for subscriptionEgressEcsToLoki.JCtx() == nil {
	//	time.Sleep(time.Second * 1)
	//}
	//// Start egress channel
	//err = process.Start(processChannel, instance.EgressSubjectEcs(), subscriptionEgressEcsToLoki.JCtx())
	//if err != nil {
	//	logger.Error().Err(err).Stack().Msg("Can't start egress stream")
	//	os.Exit(1)
	//}
	// Handle interrupt and send ping to nats
	for {
		select {
		case exit := <-c:
			logger.Info().Msgf("Received interrupt of type %v", exit)
			err := dialer.Disconnect()
			if err != nil {
				logger.Error().Err(err).Stack().Msg("Can't disconnect from nats")
			}
			if processChannel != nil {
				close(processChannel)
			}
			if lokiShipper != nil {
				lokiShipper.DisConnect()
			}
			os.Exit(1)
		case <-time.After(time.Second * 1):
			err := dialer.SendPing()
			if err != nil {
				logger.Error().Err(err).Stack().Msg("Can't send ping")
			}
			continue
		}
	}
}

func grokTest() {
	//factory, err := internalPatterns.Initialize()
	//if err != nil {
	//	panic(err)
	//}

	//{ // compile once
	//	if err != nil {
	//		panic(err)
	//	}
	//	start := time.Now()
	//	var wg sync.WaitGroup
	//	for i := 0; i < 100_000; i++ {
	//		go func(i int) {
	//			wg.Add(1)
	//			//fmt.Println(fmt.Sprintf("%d", i))
	//			_, _ = factory.Parse("COMMONAPACHELOG", fmt.Sprintf(`127.0.0.1 - - [23/Apr/2014:22:58:32 +0200] "GET /index.php HTTP/1.1" 404 %d`, i))
	//			wg.Done()
	//		}(i)
	//		//for k, v := range parseString {
	//		//	fmt.Printf("%+15s: %s\n", k, v)
	//		//}
	//	}
	//	startWait := time.Now()
	//	wg.Wait()
	//	elapsedWait := time.Since(startWait)
	//	elapsed := time.Since(start)
	//	log.Printf("Pre wait tooks %s", elapsedWait)
	//	log.Printf("Pre compiled tooks %s", elapsed)
	//}

	//{
	//	parse, err := factory.Parse("NOMAD_LOG", "2022-12-08T12:21:02.594Z [ERROR] nomad.autopilot: 😜 Failed\nto reconcile current state with the desired state\nthird line mf\n1\n3")
	//	//parse, err := factory.Parse("NOMAD_LOG", " 2022-12-08T12:21:02.594Z [ERROR] 😜")
	//	if err != nil {
	//		panic(err)
	//	}
	//	fmt.Println("---------- NOMAD_LOG ----------------")
	//	for k, v := range parse {
	//		fmt.Printf("%+15s: %s\n", k, v)
	//	}
	//	//t, err := time.Parse("2006-01-02T15:04:05.000Z", "2022-12-08T12:21:02.594Z")
	//	//if err != nil {
	//	//	panic(err)
	//	//}
	//	//fmt.Printf("Parsed date is: %v", t)
	//}
	//
	//{
	//	parse, err := factory.Parse("MSG", "pam_unix(sudo:session): session opened for user root(uid=0)\n by cloudmaster(uid=1001)")
	//	if err != nil {
	//		panic(err)
	//	}
	//	fmt.Println("---------- MSG_LOG  ----------------")
	//	for k, v := range parse {
	//		fmt.Printf("%+15s: %s\n", k, v)
	//	}
	//}
}
