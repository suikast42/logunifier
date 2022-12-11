## Jetstrem subsribers  
    // Process single message infight 
    // Suiatable approcach where the order of mesasages is bery important like IOT data
    //
	subOpts := []nats.SubOpt{
		nats.BindStream(r.streamName),
		nats.AckWait(time.Second), // Redeliver after
		//nats.MaxDeliver(5),  // Redeliver max default is infinite
		nats.ManualAck(),      // Control the ack inProgress and nack self
		nats.ReplayInstant(),  // Replay so fast as possible
		nats.DeliverAll(),     // Redeliver all not acked when restarted
		nats.MaxAckPending(1), // Max inflight ack
		nats.EnableFlowControl(), // Flow control messages enable JetStream to dynamically attempt to determine an optimal delivery message rate for the client
		nats.IdleHeartbeat(time.Second * 1), // Needed for FlowControl
		nats.Durable(r.durableSubscriptionName),
	}
	subscribe, err := js.Subscribe("", func(msg *nats.Msg) {
		//logger.Info().Msgf("%s %s", v.Timestamp, v.Message)
		logger.Info().Msgf("%s", msg.Data)
		//msg.Ack()
		//msg.Nak()
        // Simulate work for 10 second 
		go func() {
			for i := 0; i < 100; i++ {
				msg.InProgress()
				time.Sleep(time.Millisecond * 100)
			}
		}()
		msg.NakWithDelay(time.Second * 10) // Redeliver after this time 
        msg.Nak() //Redeliver as soon as possible
        msg.Term() // Term tells the server to not redeliver this message, regardless of the value
	}, subOpts...)

