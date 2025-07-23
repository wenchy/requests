package requests

import (
	"context"
	"net/http"
	"net/http/httputil"
	"time"
)

// ClientOption is the functional option type.
type ClientOption func(*Client)

// WithTimeout specifies a time limit for each client request.
//
// A Timeout of zero means no timeout. Default is zero.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.client.Timeout = timeout
	}
}

// WithTransport specifies a transport for client.
func WithTransport(transport http.RoundTripper) ClientOption {
	return func(c *Client) {
		c.client.Transport = transport
	}
}

// WithInterceptor specifies an interceptor for client.
// You can use [ChainInterceptors] to chain multiple interceptors into one.
func WithInterceptor(interceptor InterceptorFunc) ClientOption {
	return func(c *Client) {
		c.interceptor = interceptor
	}
}

// Client is an HTTP client which wraps around [http.Client] for elegant APIs and easy use.
type Client struct {
	client      *http.Client
	interceptor InterceptorFunc
}

// NewClient creates a new client to serve HTTP requests.
func NewClient(setters ...ClientOption) *Client {
	client := newDefaultClient()
	for _, setter := range setters {
		setter(client)
	}
	return client
}

// request is the common func to send an HTTP request.
func (c *Client) request(method, url string, opts *Options, body []byte) (*Response, error) {
	r, err := newRequest(method, url, opts, body)
	if err != nil {
		return nil, err
	}
	if r.opts.DumpRequestOut != nil {
		reqDump, err := httputil.DumpRequestOut(r.Request, true)
		if err != nil {
			return nil, err
		}
		*r.opts.DumpRequestOut = string(reqDump)
	}
	ctx := opts.ctx
	if opts.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}
	var interceptors []InterceptorFunc
	if r.opts.Interceptor != nil {
		interceptors = append(interceptors, r.opts.Interceptor)
	}
	if c.interceptor != nil {
		interceptors = append(interceptors, c.interceptor)
	}
	interceptor := ChainInterceptors(interceptors...)
	if interceptor != nil {
		return interceptor(ctx, r, c.do)
	}
	return c.do(ctx, r)
}

// do sends an HTTP request and returns an HTTP response, following policy
// (such as redirects, cookies, auth) as configured on the client.
func (c *Client) do(ctx context.Context, r *Request) (*Response, error) {
	// If the returned error is nil, the Response will contain
	// a non-nil Body which the user is expected to close.
	resp, err := c.client.Do(r.Request)
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

// Get sends an HTTP request with GET method.
//
// On error, any Response can be ignored. A non-nil Response with a
// non-nil error only occurs when Response.StatusCode() is not 2xx.
func (c *Client) Get(url string, options ...Option) (*Response, error) {
	return c.callMethod(http.MethodGet, url, options...)
}

// Post sends an HTTP POST request.
func (c *Client) Post(url string, options ...Option) (*Response, error) {
	return c.callMethod(http.MethodPost, url, options...)
}

// Put sends an HTTP request with PUT method.
//
// On error, any Response can be ignored. A non-nil Response with a
// non-nil error only occurs when Response.StatusCode() is not 2xx.
func (c *Client) Put(url string, options ...Option) (*Response, error) {
	return c.callMethod(http.MethodPut, url, options...)
}

// Patch sends an HTTP request with PATCH method.
//
// On error, any Response can be ignored. A non-nil Response with a
// non-nil error only occurs when Response.StatusCode() is not 2xx.
func (c *Client) Patch(url string, options ...Option) (*Response, error) {
	return c.callMethod(http.MethodPatch, url, options...)
}

// Delete sends an HTTP request with DELETE method.
//
// On error, any Response can be ignored. A non-nil Response with a
// non-nil error only occurs when Response.StatusCode() is not 2xx.
func (c *Client) Delete(url string, options ...Option) (*Response, error) {
	return c.callMethod(http.MethodDelete, url, options...)
}

func (c *Client) callMethod(method, url string, options ...Option) (*Response, error) {
	opts := parseOptions(options...)
	return dispatchers[opts.bodyType](c, method, url, opts)
}
