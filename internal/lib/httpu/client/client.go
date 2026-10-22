package client

import (
	"net/http"
	"time"
)

type config struct {
	timeout             time.Duration
	maxIdleConns        int
	maxConnsPerHosts    int
	maxIdleConnsPerHost int
}

type Option func(*config)

func WithMaxIdleConnx(val int) Option {
	return func(c *config) {
		c.maxIdleConns = val
	}
}

func WithMaxConnsPerHosts(val int) Option {
	return func(c *config) {
		c.maxConnsPerHosts = val
	}
}

func WithMaxIdleConnsPerHost(val int) Option {
	return func(c *config) {
		c.maxIdleConnsPerHost = val
	}
}

func WithTimeout(val time.Duration) Option {
	return func(c *config) {
		c.timeout = val
	}
}

func NewHTTPClient(opts ...Option) *http.Client {
	c := &config{
		timeout:             10 * time.Second,
		maxIdleConns:        100,
		maxConnsPerHosts:    100,
		maxIdleConnsPerHost: 100,
	}

	for _, opt := range opts {
		opt(c)
	}

	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = c.maxIdleConns
	t.MaxConnsPerHost = c.maxConnsPerHosts
	t.MaxIdleConnsPerHost = c.maxIdleConnsPerHost

	return &http.Client{
		Timeout:   c.timeout,
		Transport: t,
	}
}
