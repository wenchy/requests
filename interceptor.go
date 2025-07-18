package requests

import (
	"context"
)

// Do is called by Interceptor to complete HTTP requests.
type Do func(ctx context.Context, r *Request) (*Response, error)

// InterceptorFunc provides a hook to intercept the execution of an HTTP request
// invocation. When an interceptor(s) is set, requests delegates all HTTP
// client invocations to the interceptor, and it is the responsibility of the
// interceptor to call do to complete the processing of the HTTP request.
type InterceptorFunc func(ctx context.Context, r *Request, do Do) (*Response, error)

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
