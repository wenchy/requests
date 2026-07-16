package requests

import (
	"context"
)

// Do is called by an interceptor to complete the HTTP request.
type Do func(ctx context.Context, r *Request) (*Response, error)

// InterceptorFunc intercepts an HTTP request. When set, the client delegates
// the request to the interceptor, which must call do to complete it.
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
