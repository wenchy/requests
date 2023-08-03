package requests

import (
	"net"
	"net/http"
	"time"
)

type Env struct {
	Timeout          time.Duration
	DefaultTransport *http.Transport
}

var env Env

func init() {
	env.Timeout = 60 * time.Second // default timeout
	// DefaultTransport
	// refer: https://github.com/golang/go/blob/c333d07ebe9268efc3cf4bd68319d65818c75966/src/net/http/transport.go#L42
	env.DefaultTransport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

func SetEnvTimeout(timeout time.Duration) {
	env.Timeout = timeout
}
