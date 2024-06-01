package requests

import (
	"context"
	"net/http"
)

// Do is called by Interceptor to complete HTTP requests.
type Do func(ctx context.Context, r *http.Request) (*http.Response, error)

// Interceptor provides a hook to intercept the execution of an HTTP request
// invocation. When an interceptor(s) is set, requests delegates all HTTP
// client invocations to the interceptor, and it is the responsibility of the
// interceptor to call do to complete the processing of the HTTP request.
type Interceptor func(ctx context.Context, r *http.Request, do Do) (*http.Response, error)

type Client struct {
	*http.Client
}

// Do sends the HTTP request and returns after response is received.
func (c *Client) Do(ctx context.Context, r *http.Request) (*http.Response, error) {
	if env.interceptor != nil {
		do := func(ctx context.Context, r *http.Request) (*http.Response, error) {
			return c.Client.Do(r)
		}
		return env.interceptor(ctx, r, do)
	}
	return c.Client.Do(r)
}
