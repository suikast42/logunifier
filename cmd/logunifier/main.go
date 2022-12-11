package main

import (
	"fmt"
	"github.com/suikast42/logunifier/internal/bootstrap"
	"github.com/suikast42/logunifier/internal/config"
	"github.com/suikast42/logunifier/internal/model/ingress/journald"
	internalPatterns "github.com/suikast42/logunifier/pkg/patterns"
)

func main() {
	config.ReadConfigs()
	err := config.ConfigLogging()
	if err != nil {
		panic(err)
	}

	instance, err := config.Instance()
	if err != nil {
		panic(err)
	}
	logger := config.Logger()
	subscriptionIngressJournald := journald.NewSubscription("IngressLogsJournaldStream", "IngressLogsJournaldProcessor", instance.IngressNatsJournald())
	subscriptions := []bootstrap.NatsSubscription{subscriptionIngressJournald}
	err = bootstrap.Connect(&subscriptions)
	if err != nil {
		logger.Error().Err(err).Stack().Msg("Can't connect to nats")
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
