package requests

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/Wenchy/requests/internal/auth"
)

// Options defines all optional parameters for HTTP request.
type Options struct {
	ctx context.Context

	Headers map[string]string
	Params  map[string]string
	// body
	Body io.Reader
	// different request body types
	Data  any
	Form  map[string]string
	JSON  any
	Files map[string]*os.File
	// different response body types
	ToText *string
	ToJSON any

	// auth
	AuthInfo *auth.AuthInfo
	// request timeout
	Timeout time.Duration

	DisableKeepAlives bool
	// dump
	DumpRequestOut *string
	DumpResponse   *string
}

// Option is the functional option type.
type Option func(*Options)

// Context sets the HTTP request context.
//
// For outgoing client request, the context controls the entire lifetime of
// a request and its response: obtaining a connection, sending the request,
// and reading the response headers and body.
func Context(ctx context.Context) Option {
	return func(opts *Options) {
		opts.ctx = ctx
	}
}

// Headers sets the HTTP headers.
func Headers(headers map[string]string) Option {
	return func(opts *Options) {
		if opts.Headers != nil {
			for k, v := range headers {
				opts.Headers[k] = v
			}
		} else {
			opts.Headers = headers
		}
	}
}

// HeaderPairs sets HTTP headers formed by the mapping of key, value ...
// Pairs panics if len(kv) is odd.
func HeaderPairs(kv ...string) Option {
	if len(kv)%2 == 1 {
		panic(fmt.Sprintf("params: got the odd number of input pairs: %d", len(kv)))
	}
	headers := map[string]string{}
	var key string
	for i, s := range kv {
		if i%2 == 0 {
			key = s
			continue
		}
		headers[key] = s
	}
	return Headers(headers)
}

// Params sets the given params into the URL querystring.
func Params(params map[string]string) Option {
	return func(opts *Options) {
		if opts.Params != nil {
			for k, v := range params {
				opts.Params[k] = v
			}
		} else {
			opts.Params = params
		}
	}
}

// ParamPairs returns an Params formed by the mapping of key, value ...
// Pairs panics if len(kv) is odd.
func ParamPairs(kv ...string) Option {
	if len(kv)%2 == 1 {
		panic(fmt.Sprintf("params: got the odd number of input pairs: %d", len(kv)))
	}
	params := map[string]string{}
	var key string
	for i, s := range kv {
		if i%2 == 0 {
			key = s
			continue
		}
		params[key] = s
	}
	return Params(params)
}

// Body sets io.Reader to hold request body.
func Body(body io.Reader) Option {
	return func(opts *Options) {
		opts.Body = body
	}
}

// Data sets raw string into the request body.
func Data(data any) Option {
	return func(opts *Options) {
		opts.Data = data
	}
}

// Form sets the given form into the request body.
// It also sets the Content-Type as "application/x-www-form-urlencoded".
func Form(form map[string]string) Option {
	return func(opts *Options) {
		opts.Form = form
	}
}

// FormPairs sets form by the mapping of key, value ...
// Pairs panics if len(kv) is odd.
func FormPairs(kv ...string) Option {
	if len(kv)%2 == 1 {
		panic(fmt.Sprintf("params: got the odd number of input pairs: %d", len(kv)))
	}
	form := map[string]string{}
	var key string
	for i, s := range kv {
		if i%2 == 0 {
			key = s
			continue
		}
		form[key] = s
	}
	return Form(form)
}

// JSON marshals the given struct as JSON into the request body.
// It also sets the Content-Type as "application/json".
func JSON(v any) Option {
	return func(opts *Options) {
		opts.JSON = v
	}
}

// ToText unmarshals HTTP response body to string.
func ToText(v *string) Option {
	return func(opts *Options) {
		opts.ToText = v
	}
}

// ToJSON unmarshals HTTP response body to given struct as JSON.
func ToJSON(v any) Option {
	return func(opts *Options) {
		opts.ToJSON = v
	}
}

// BasicAuth is the option to implement HTTP Basic Auth.
func BasicAuth(username, password string) Option {
	return func(opts *Options) {
		opts.AuthInfo = &auth.AuthInfo{
			Type:     auth.BasicAuth,
			Username: username,
			Password: password,
		}
	}
}

// Files sets files to a map of (field, fileHandler).
// It also sets the Content-Type as "multipart/form-data".
func Files(files map[string]*os.File) Option {
	return func(opts *Options) {
		if opts.Files != nil {
			for k, v := range files {
				opts.Files[k] = v
			}
		} else {
			opts.Files = files
		}
	}
}

// Timeout specifies a time limit for requests made by this
// Client. The timeout includes connection time, any
// redirects, and reading the response body. The timer remains
// running after Get, Head, Post, or Do return and will
// interrupt reading of the Response.Body.
//
// A Timeout of zero means no timeout. Default is 60s.
func Timeout(timeout time.Duration) Option {
	return func(opts *Options) {
		opts.Timeout = timeout
	}
}

// DisableKeepAlives, if true, disables HTTP keep-alives and will
// only use the connection to the server for a single HTTP request.
//
// This is unrelated to the similarly named TCP keep-alives.
func DisableKeepAlives() Option {
	return func(opts *Options) {
		opts.DisableKeepAlives = true
	}
}

// Dump dumps outgoing client request and response to the corresponding
// input param (req or resp) if not nil.
//
// Refer:
// - https://pkg.go.dev/net/http/httputil#DumpRequestOut
// - https://pkg.go.dev/net/http/httputil#DumpResponse
func Dump(req, resp *string) Option {
	return func(opts *Options) {
		opts.DumpRequestOut = req
		opts.DumpResponse = resp
	}
}

// newDefaultOptions creates a new default HTTP options.
func newDefaultOptions() *Options {
	return &Options{
		Headers: map[string]string{},
		Params:  map[string]string{},
		Form:    nil,
		JSON:    nil,
		Timeout: env.timeout,
	}
}

func parseOptions(options ...Option) *Options {
	opts := newDefaultOptions()
	for _, setter := range options {
		setter(opts)
	}
	return opts
}
