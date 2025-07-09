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
	transport *http.Transport
	// hostRoundTrippers specify the host-specific RoundTripper to use for the
	// request. If not found, http.DefaultTransport is used.
	hostRoundTrippers map[string]http.RoundTripper
	// interceptor intercepts each HTTP request.
	interceptor InterceptorFunc
}

var env environment

func init() {
	env.timeout = 60 * time.Second // default timeout
	transport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		panic("Ooh! http.DefaultTransport's underlying is not *http.Transport. Maybe golang team has changed it.")
	}
	env.transport = transport
}

// SetEnvTimeout sets the default timeout for each HTTP request at
// the environment level.
func SetEnvTimeout(timeout time.Duration) {
	env.timeout = timeout
}

// WithInterceptor specifies the interceptor for each HTTP request.
func WithInterceptor(interceptors ...InterceptorFunc) {
	// Prepend env.interceptor to the chaining interceptors if it exists, since
	// env.interceptor will be executed before any other chained interceptors.
	if env.interceptor != nil {
		interceptors = append([]InterceptorFunc{env.interceptor}, interceptors...)
	}
	env.interceptor = ChainInterceptors(interceptors...)
}

// ChainInterceptors chains multiple interceptors into one.
func ChainInterceptors(interceptors ...InterceptorFunc) InterceptorFunc {
	switch len(interceptors) {
	case 0:
		return nil
	case 1:
		return interceptors[0]
	default:
		return func(ctx context.Context, r *Request, do Do) (*Response, error) {
			return interceptors[0](ctx, r, getChainDo(interceptors, 0, do))
		}
	}
}

// getChainDo generates the chained do recursively.
func getChainDo(interceptors []InterceptorFunc, curr int, finalDo Do) Do {
	if curr == len(interceptors)-1 {
		return finalDo
	}
	return func(ctx context.Context, r *Request) (*Response, error) {
		return interceptors[curr+1](ctx, r, getChainDo(interceptors, curr+1, finalDo))
	}
}

// SetHostTransport sets the host-specific RoundTripper to use for the request.
//
// # Example
//
//	SetHostTransport(map[string]http.RoundTripper{
//	    "example1.com": http.DefaultTransport,
//	    "example2.com": MyCustomTransport,
//	})
func SetHostTransport(rts map[string]http.RoundTripper) {
	env.hostRoundTrippers = rts
}
