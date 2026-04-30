package httpclient

import (
	"net"
	"net/http"
	"time"

	"go-far/internal/preference"

	"github.com/failsafe-go/failsafe-go"
	"github.com/failsafe-go/failsafe-go/circuitbreaker"
	"github.com/failsafe-go/failsafe-go/failsafehttp"
	"github.com/failsafe-go/failsafe-go/retrypolicy"
	"github.com/failsafe-go/failsafe-go/timeout"
	"github.com/rs/zerolog"
)

// HttpClientOptions holds HTTP client configuration
type HttpClientOptions struct {
	CircuitBreaker        CircuitBreakerOptions `yaml:"circuit_breaker"`
	KeepAlive             time.Duration         `yaml:"keep_alive"`
	MaxIdleConnsPerHost   int                   `yaml:"max_idle_conns_per_host"`
	MaxConnsPerHost       int                   `yaml:"max_conns_per_host"`
	IdleConnTimeout       time.Duration         `yaml:"idle_conn_timeout"`
	DialTimeout           time.Duration         `yaml:"dial_timeout"`
	TLSHandshakeTimeout   time.Duration         `yaml:"tls_handshake_timeout"`
	ResponseHeaderTimeout time.Duration         `yaml:"response_header_timeout"`
	ExpectContinueTimeout time.Duration         `yaml:"expect_continue_timeout"`
	Timeout               time.Duration         `yaml:"timeout"`
	MaxIdleConns          int                   `yaml:"max_idle_conns"`
	Enabled               bool                  `yaml:"enabled"`
	DisableCompression    bool                  `yaml:"disable_compression"`
}

// CircuitBreakerOptions holds circuit breaker configuration
type CircuitBreakerOptions struct {
	MaxRetries int           `yaml:"max_retries"`
	BackoffMin time.Duration `yaml:"backoff_min"`
	BackoffMax time.Duration `yaml:"backoff_max"`
}

// getTraceInfo extracts trace/span ID from failsafe execution event context
func getTraceInfo(e failsafe.ExecutionInfo) (traceID, spanID string) {
	ctx := e.Context()
	if ctx == nil {
		return "", ""
	}

	if v, ok := ctx.Value(preference.CONTEXT_KEY_LOG_TRACE_ID).(string); ok && v != "" {
		traceID = v
	}

	if v, ok := ctx.Value(preference.CONTEXT_KEY_LOG_SPAN_ID).(string); ok && v != "" {
		spanID = v
	}

	return traceID, spanID
}

// createTimeoutPolicy creates a timeout policy with tracing support
func createTimeoutPolicy(log *zerolog.Logger) failsafe.Policy[*http.Response] {
	return timeout.NewBuilder[*http.Response](3 * time.Second).
		OnTimeoutExceeded(func(e failsafe.ExecutionDoneEvent[*http.Response]) {
			event := log.Info()
			if traceID, spanID := getTraceInfo(e); traceID != "" || spanID != "" {
				if traceID != "" {
					event = event.Str(string(preference.CONTEXT_KEY_LOG_TRACE_ID), traceID)
				}

				if spanID != "" {
					event = event.Str(string(preference.CONTEXT_KEY_LOG_SPAN_ID), spanID)
				}
			}
			event.Msg("Request timed out")
		}).Build()
}

// createRetryPolicy creates a retry policy with tracing support
func createRetryPolicy(log *zerolog.Logger, opt *CircuitBreakerOptions) failsafe.Policy[*http.Response] {
	return retrypolicy.NewBuilder[*http.Response]().
		WithMaxRetries(opt.MaxRetries).
		WithBackoff(opt.BackoffMin, opt.BackoffMax).
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
			event := log.Info().Int("attempt", e.Attempts()).Dur("delay", e.Delay)
			if traceID, spanID := getTraceInfo(e); traceID != "" || spanID != "" {
				if traceID != "" {
					event = event.Str(string(preference.CONTEXT_KEY_LOG_TRACE_ID), traceID)
				}

				if spanID != "" {
					event = event.Str(string(preference.CONTEXT_KEY_LOG_SPAN_ID), spanID)
				}
			}
			event.Msg("Retry scheduled")
		}).Build()
}

// createCircuitBreakerPolicy creates a circuit breaker policy
func createCircuitBreakerPolicy(log *zerolog.Logger) failsafe.Policy[*http.Response] {
	return circuitbreaker.NewBuilder[*http.Response]().
		HandleIf(func(response *http.Response, err error) bool {
			return response != nil && response.StatusCode == http.StatusServiceUnavailable
		}).
		WithDelayFunc(failsafehttp.DelayFunc).
		OnStateChanged(func(event circuitbreaker.StateChangedEvent) {
			log.Info().
				Str("old_state", event.OldState.String()).
				Str("new_state", event.NewState.String()).
				Msg("Circuit breaker state changed")
		}).Build()
}

func InitHttpClient(log *zerolog.Logger, opt *HttpClientOptions) *http.Client {
	if !opt.Enabled {
		log.Debug().Msg("HTTP client is disabled")
		return nil
	}

	baseTransport := &http.Transport{
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

	transport := baseTransport
	timeoutPolicy := createTimeoutPolicy(log)
	circuitBreakerPolicy := createCircuitBreakerPolicy(log)
	retryPolicy := createRetryPolicy(log, &opt.CircuitBreaker)

	wrappedTransport := failsafehttp.NewRoundTripper(
		transport,            // innerRoundTripper
		circuitBreakerPolicy, // circuit breaker
		retryPolicy,          // retry policy
		timeoutPolicy,        // timeout policy
	)

	client := &http.Client{
		Transport: wrappedTransport,
		Timeout:   opt.Timeout,
	}

	log.Info().Msg("HTTP client initialized")

	return client
}
