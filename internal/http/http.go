package http

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type transport struct {
	http.RoundTripper
	token string
}

// RoundTrip handles Put.io authentication and tracing requests.
func (t transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.RoundTripper == nil {
		t.RoundTripper = http.DefaultTransport // for safety reason
	}

	// log requests
	logger := log.FromContext(req.Context(), "method", req.Method, "url", req.URL.String())
	logger.WithName("http")
	logger.Info("URL being requested")

	// insert token
	req.Header.Set("Authorization", "Bearer "+t.token)
	return t.RoundTripper.RoundTrip(req) //nolint:wrapcheck
}

func NewHTTPClient(token string) *http.Client {
	return &http.Client{Transport: &transport{
		RoundTripper: otelhttp.NewTransport(nil), // trace requests
		token:        token,                      // insert token on each requests
	}}
}
