package requests

import (
	"context"
	"net/http"
	"net/http/httputil"
)

// Do is called by Interceptor to complete HTTP requests.
type Do func(ctx context.Context, r *Request) (*Response, error)

// Interceptor provides a hook to intercept the execution of an HTTP request
// invocation. When an interceptor(s) is set, requests delegates all HTTP
// client invocations to the interceptor, and it is the responsibility of the
// interceptor to call do to complete the processing of the HTTP request.
type Interceptor func(ctx context.Context, r *Request, do Do) (*Response, error)

type Client struct {
	*http.Client
}

// Do sends the HTTP request and returns after response is received.
func (c *Client) Do(ctx context.Context, r *Request) (*Response, error) {
	if r.opts.DumpRequestOut != nil {
		reqDump, err := httputil.DumpRequestOut(r.Request, true)
		if err != nil {
			return nil, err
		}
		*r.opts.DumpRequestOut = string(reqDump)
	}
	if ctx != nil {
		r = r.WithContext(ctx)
	}
	if env.interceptor != nil {
		return env.interceptor(ctx, r, c.do)
	}
	return c.do(ctx, r)
}

func (c *Client) do(ctx context.Context, r *Request) (*Response, error) {
	// If the returned error is nil, the Response will contain
	// a non-nil Body which the user is expected to close.
	resp, err := c.Client.Do(r.Request)
	if err != nil {
		return nil, err
	}
	if r.opts.DumpResponse != nil {
		respDump, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return nil, err
		}
		*r.opts.DumpResponse = string(respDump)
	}

	return newResponse(resp, r.opts)
}
