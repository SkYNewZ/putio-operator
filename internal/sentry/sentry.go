package sentry

import (
	"fmt"
	"os"

	"github.com/getsentry/sentry-go"
)

// ConfigureSentry make a new sentry.Client.
func ConfigureSentry(serviceName, serviceVersion string) (*sentry.Client, error) {
	client, err := sentry.NewClient(sentry.ClientOptions{
		// set the SENTRY_DSN environment variable.
		Dsn: "",

		// set the SENTRY_ENVIRONMENT
		Environment: "",

		Release: serviceName + "@" + serviceVersion,

		// Enable printing of SDK debug messages.
		// Useful when getting started or trying to figure something out.
		Debug: os.Getenv("DEBUG") == "1",

		// Set TracesSampleRate to 1.0 to capture 100%
		// of transactions for performance monitoring.
		// We recommend adjusting this value in production,
		TracesSampleRate: 1.0,
	})
	if err != nil {
		return nil, fmt.Errorf("sentry: failed to configure sentry client: %w", err)
	}

	return client, nil
}
