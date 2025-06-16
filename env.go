package requests

import (
	"context"
	"net/http"
	"time"
)

type environment struct {
	timeout time.Duration
	// transport establishes network connections as needed and
	// caches them for reuse by subsequent calls. It uses HTTP proxies as
	// directed by the environment variables HTTP_PROXY, HTTPS_PROXY and
	// NO_PROXY (or the lowercase versions thereof).
	transport   *http.Transport
	interceptor Interceptor
}

var env environment

func init() {
	env.timeout = 60 * time.Second                                  // default timeout
	env.transport = http.DefaultTransport.(*http.Transport).Clone() // default transport
}

// SetEnvTimeout sets the default timeout for each HTTP request at
// the environment level.
func SetEnvTimeout(timeout time.Duration) {
	env.timeout = timeout
}

// WithInterceptor specifies the interceptor for each HTTP request.
func WithInterceptor(interceptors ...Interceptor) {
	// Prepend env.interceptor to the chaining interceptors if it exists, since
	// env.interceptor will be executed before any other chained interceptors.
	if env.interceptor != nil {
		interceptors = append([]Interceptor{env.interceptor}, interceptors...)
	}
	var chainedInt Interceptor
	if len(interceptors) == 0 {
		chainedInt = nil
	} else if len(interceptors) == 1 {
		chainedInt = interceptors[0]
	} else {
		chainedInt = func(ctx context.Context, r *Request, do Do) (*Response, error) {
			return interceptors[0](ctx, r, getChainDo(interceptors, 0, do))
		}
	}
	env.interceptor = chainedInt
}

// getChainDo recursively generates the chained do.
func getChainDo(interceptors []Interceptor, curr int, finalDo Do) Do {
	if curr == len(interceptors)-1 {
		return finalDo
	}
	return func(ctx context.Context, r *Request) (*Response, error) {
		return interceptors[curr+1](ctx, r, getChainDo(interceptors, curr+1, finalDo))
	}
}
