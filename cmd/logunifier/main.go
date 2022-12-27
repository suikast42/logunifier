package main

import (
	"fmt"
	"github.com/suikast42/logunifier/internal/bootstrap"
	"github.com/suikast42/logunifier/internal/config"
	"github.com/suikast42/logunifier/internal/streams/egress"
	"github.com/suikast42/logunifier/internal/streams/ingress/testingress"
	internalPatterns "github.com/suikast42/logunifier/pkg/patterns"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	//var validationChannel  = chan *egress.MsgContext

	//ctx, cancelFunc := context.WithCancel(context.Background())
	// Listen on os exit signals
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	//Read the config file -> program flags
	config.ReadConfigs()
	err := config.ConfigLogging()
	// The first panic point. Must be sure that the config is done
	if err != nil {
		panic(err)
	}

	instance, err := config.Instance()
	// The second  panic point. Must be sure that the config is done
	if err != nil {
		panic(err)
	}
	logger := config.Logger()
	processChannel := make(chan *egress.MsgContext, 4096)

	//TODO find a nicer way to define the JetStream subscriptions
	//subscriptionIngressJournald, err := journald.NewSubscription("IngressLogsJournaldStream", "IngressLogsJournaldProcessor", []string{instance.IngressNatsJournald()}, processChannel)
	//if err != nil {
	//	logger.Error().Err(err).Msgf("Can't subscribe to %v", instance.IngressNatsJournald())
	//	os.Exit(1)
	//}

	subscriptionIngressTest, err := testingress.NewSubscription("IngressLogsTestStream", "IngressLogsTestStreamProcessor", []string{instance.IngresNatsTest()}, processChannel)
	if err != nil {
		logger.Error().Err(err).Msgf("Can't subscribe to %v", instance.IngresNatsTest())
		os.Exit(1)
	}

	subscriptions := []bootstrap.NatsSubscription{
		//subscriptionIngressJournald,
		subscriptionIngressTest,
	}

	// Connect to nats server(s) should be a save background process
	// It handles reconnection logic itself
	// If an error occurs here then something is general wrong.
	// Exit here
	dialer := bootstrap.New(subscriptions)
	err = dialer.Connect()
	if err != nil {
		logger.Error().Err(err).Stack().Msg("Can't connect to nats")
		os.Exit(1)
	}
	// Start egress channel
	err = egress.Start(processChannel, instance.EgressSubject(), dialer.Connection())
	if err != nil {
		logger.Error().Err(err).Stack().Msg("Can't start egress stream")
		os.Exit(1)
	}
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
	factory, err := internalPatterns.New()
	if err != nil {
		panic(err)
	}

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

	{
		parse, err := factory.Parse("NOMAD_LOG", "2022-12-08T12:21:02.594Z [ERROR] nomad.autopilot: 😜 Failed\nto reconcile current state with the desired state\nthird line mf\n1\n3")
		//parse, err := factory.Parse("NOMAD_LOG", " 2022-12-08T12:21:02.594Z [ERROR] 😜")
		if err != nil {
			panic(err)
		}
		fmt.Println("---------- NOMAD_LOG ----------------")
		for k, v := range parse {
			fmt.Printf("%+15s: %s\n", k, v)
		}
		//t, err := time.Parse("2006-01-02T15:04:05.000Z", "2022-12-08T12:21:02.594Z")
		//if err != nil {
		//	panic(err)
		//}
		//fmt.Printf("Parsed date is: %v", t)
	}

	{
		parse, err := factory.Parse("MSG", "pam_unix(sudo:session): session opened for user root(uid=0)\n by cloudmaster(uid=1001)")
		if err != nil {
			panic(err)
		}
		fmt.Println("---------- MSG_LOG  ----------------")
		for k, v := range parse {
			fmt.Printf("%+15s: %s\n", k, v)
		}
	}
}
