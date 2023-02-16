package health

import (
	"context"
	"fmt"
	"github.com/alexliesenfeld/health"
	"github.com/suikast42/logunifier/internal/bootstrap"
	"github.com/suikast42/logunifier/internal/config"
	"github.com/suikast42/logunifier/internal/streams/connectors/lokishipper"
	"net/http"
	"time"
)

func Start(path string, port int, dialer *bootstrap.NatsDialer, lokishipper *lokishipper.LokiShipper) error {
	logger := config.Logger()
	checker := health.NewChecker(

		// Set the time-to-live for our cache to 1 second (default).
		health.WithCacheDuration(1*time.Second),

		// Configure a global timeout that will be applied to all checks.
		health.WithTimeout(10*time.Second),

		// A check configuration to see if our database connection is up.
		// The check function will be executed for each HTTP request.
		health.WithCheck(health.Check{
			Name:    "nats",          // A unique check name.
			Timeout: 2 * time.Second, // A check specific timeout.
			Check:   dialer.Health,
		}),
		health.WithCheck(health.Check{
			Name:    "loki",          // A unique check name.
			Timeout: 2 * time.Second, // A check specific timeout.
			Check:   lokishipper.Health,
		}),
		// Set a status listener that will be invoked when the health status changes.
		// More powerful hooks are also available (see docs).
		health.WithStatusListener(func(ctx context.Context, state health.CheckerState) {
			logger.Info().Msgf("health status changed to %s", state.Status)
		}),
	)

	// Create a new health check http.Handler that returns the health status
	// serialized as a JSON string. You can pass pass further configuration
	// options to NewHandler to modify default configuration.
	http.Handle(path, health.NewHandler(checker))
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		return err
	}

	return nil
}
