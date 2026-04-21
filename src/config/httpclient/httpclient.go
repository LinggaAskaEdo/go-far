package httpclient

import (
	"net"
	"net/http"
	"time"

	"github.com/failsafe-go/failsafe-go"
	"github.com/failsafe-go/failsafe-go/circuitbreaker"
	"github.com/failsafe-go/failsafe-go/failsafehttp"
	"github.com/failsafe-go/failsafe-go/retrypolicy"
	"github.com/failsafe-go/failsafe-go/timeout"
	"github.com/rs/zerolog"
)

type HttpClientOptions struct {
	Enabled               bool                  `yaml:"enabled"`
	MaxIdleConns          int                   `yaml:"max_idle_conns"`
	MaxIdleConnsPerHost   int                   `yaml:"max_idle_conns_per_host"`
	MaxConnsPerHost       int                   `yaml:"max_conns_per_host"`
	IdleConnTimeout       time.Duration         `yaml:"idle_conn_timeout"`
	DialTimeout           time.Duration         `yaml:"dial_timeout"`
	KeepAlive             time.Duration         `yaml:"keep_alive"`
	TLSHandshakeTimeout   time.Duration         `yaml:"tls_handshake_timeout"`
	ResponseHeaderTimeout time.Duration         `yaml:"response_header_timeout"`
	ExpectContinueTimeout time.Duration         `yaml:"expect_continue_timeout"`
	DisableCompression    bool                  `yaml:"disable_compression"`
	Timeout               time.Duration         `yaml:"timeout"`
	CircuitBreaker        CircuitBreakerOptions `yaml:"circuit_breaker"`
}

type CircuitBreakerOptions struct {
	MaxRetries int           `yaml:"max_retries"`
	BackoffMin time.Duration `yaml:"backoff_min"`
	BackoffMax time.Duration `yaml:"backoff_max"`
}

func InitHttpClient(log zerolog.Logger, opt HttpClientOptions) *http.Client {
	if !opt.Enabled {
		log.Debug().Msg("HTTP client is disabled")
		return nil
	}

	transport := &http.Transport{
		MaxIdleConns:          opt.MaxIdleConns,
		MaxIdleConnsPerHost:   opt.MaxIdleConnsPerHost,
		MaxConnsPerHost:       opt.MaxConnsPerHost,
		IdleConnTimeout:       opt.IdleConnTimeout,
		DisableCompression:    opt.DisableCompression,
		TLSHandshakeTimeout:   opt.TLSHandshakeTimeout,
		ResponseHeaderTimeout: opt.ResponseHeaderTimeout,
		ExpectContinueTimeout: opt.ExpectContinueTimeout,
		DialContext: (&net.Dialer{
			Timeout:   opt.DialTimeout,
			KeepAlive: opt.KeepAlive,
		}).DialContext,
	}

	timeout := timeout.NewBuilder[*http.Response](3 * time.Second).
		OnTimeoutExceeded(func(e failsafe.ExecutionDoneEvent[*http.Response]) {
			log.Info().Msg("Request timed out")
		}).Build()

	circuitBreaker := circuitbreaker.NewBuilder[*http.Response]().
		HandleIf(func(response *http.Response, err error) bool {
			return response != nil && response.StatusCode == http.StatusServiceUnavailable
		}).
		WithDelayFunc(failsafehttp.DelayFunc).
		OnStateChanged(func(event circuitbreaker.StateChangedEvent) {
			log.Info().Str("old_state", event.OldState.String()).Str("new_state", event.NewState.String()).Msg("Circuit breaker state changed")
		}).
		Build()

	retryPolicy := retrypolicy.NewBuilder[*http.Response]().
		WithMaxRetries(opt.CircuitBreaker.MaxRetries).
		WithBackoff(opt.CircuitBreaker.BackoffMin, opt.CircuitBreaker.BackoffMax).
		HandleIf(func(response *http.Response, err error) bool {
			if response == nil {
				log.Warn().Err(err).Msg("Retry attempt: response is nil")
				return true
			}

			if response.StatusCode == http.StatusServiceUnavailable {
				log.Warn().Err(err).Msg("Retry attempt: response status is 503 Service Unavailable")
				return true
			}

			return false
		}).
		OnRetryScheduled(func(e failsafe.ExecutionScheduledEvent[*http.Response]) {
			log.Info().Int("attempt", e.Attempts()).Dur("delay", e.Delay).Msg("Retry scheduled")
		}).
		Build()

	wrappedTransport := failsafehttp.NewRoundTripper(
		transport,      // innerRoundTripper
		circuitBreaker, // circuit breaker
		retryPolicy,    // retry policy
		timeout,        // timeout policy
	)

	client := &http.Client{
		Transport: wrappedTransport,
		Timeout:   opt.Timeout,
	}

	log.Info().Msg("HTTP client initialized")

	return client
}
