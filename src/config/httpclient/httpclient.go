package httpclient

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

type HttpClientOptions struct {
	Enabled               bool          `yaml:"enabled"`
	MaxIdleConns          int           `yaml:"max_idle_conns"`
	MaxIdleConnsPerHost   int           `yaml:"max_idle_conns_per_host"`
	MaxConnsPerHost       int           `yaml:"max_conns_per_host"`
	IdleConnTimeout       time.Duration `yaml:"idle_conn_timeout"`
	DialTimeout           time.Duration `yaml:"dial_timeout"`
	KeepAlive             time.Duration `yaml:"keep_alive"`
	TLSHandshakeTimeout   time.Duration `yaml:"tls_handshake_timeout"`
	ResponseHeaderTimeout time.Duration `yaml:"response_header_timeout"`
	ExpectContinueTimeout time.Duration `yaml:"expect_continue_timeout"`
	DisableCompression    bool          `yaml:"disable_compression"`
	Timeout               time.Duration `yaml:"timeout"`
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

	client := &http.Client{
		Transport: transport,
		Timeout:   opt.Timeout,
	}

	log.Debug().
		Int("max_idle_conns", opt.MaxIdleConns).
		Int("max_idle_conns_per_host", opt.MaxIdleConnsPerHost).
		Int("max_conns_per_host", opt.MaxConnsPerHost).
		Msg("HTTP client initialized")

	return client
}

func Ping(ctx context.Context, client *http.Client, url string) error {
	if client == nil {
		return nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return nil
	}

	return nil
}
