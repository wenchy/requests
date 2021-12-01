package requests

import (
	"fmt"
	"io"
	"os"
)

// Options follow the design of Functional Options(https://github.com/tmrts/go-patterns/blob/master/idiom/functional-options.md)
type Options struct {
	Headers map[string]string
	Params  map[string]string
	// body
	Body io.Reader
	// different body types
	Data  interface{}
	Form  map[string]string
	JSON  interface{}
	Files map[string]*os.File

	// auth
	Auth Auth
	// timeout seconds
	Timeout int64

	DisableKeepAlives bool
}

// Option is the functional option type.
type Option func(*Options)

// Headers set the HTTP header KVs.
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

// HeaderPairs returns an Headers formed by the mapping of key, value ... Pairs panics if len(kv) is odd.
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

// Params encode the given KV into the URL querystring.
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

// ParamPairs returns an Params formed by the mapping of key, value ... Pairs panics if len(kv) is odd.
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

// Body set io.Reader to hold request body.
func Body(body io.Reader) Option {
	return func(opts *Options) {
		opts.Body = body
	}
}

// Data set raw string into the request body.
func Data(data interface{}) Option {
	return func(opts *Options) {
		opts.Data = data
	}
}

// Form encode the given KV into the request body.
// It also sets the Content-Type as "application/x-www-form-urlencoded".
func Form(form map[string]string) Option {
	return func(opts *Options) {
		opts.Form = form
	}
}

// FormPairs returns an Form formed by the mapping of key, value ... Pairs panics if len(kv) is odd.
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

// JSON serializes the given struct as JSON into the request body.
// It also sets the Content-Type as "application/json".
func JSON(obj interface{}) Option {
	return func(opts *Options) {
		opts.JSON = obj
	}
}

// BasicAuth is the option to implement HTTP Basic Auth.
func BasicAuth(username, password string) Option {
	return func(opts *Options) {
		opts.Auth = Auth{
			authType: HTTPBasicAuth,
			username: username,
			password: password,
		}
	}
}

// Files sets files to a map of (field, fileHandler).
// It also sets the Content-Type as "multipart/form-data"
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
func Timeout(timeout int64) Option {
	return func(opts *Options) {
		opts.Timeout = timeout
	}
}

func DisableKeepAlives() Option {
	return func(opts *Options) {
		opts.DisableKeepAlives = true
	}
}

func newDefaultOptions() *Options {
	return &Options{
		Headers: map[string]string{},
		Params:  map[string]string{},
		Form:    nil,
		JSON:    nil,
		Timeout: env.Timeout,
	}
}

func parseOptions(setters ...Option) *Options {
	// Default Options
	opts := newDefaultOptions()
	for _, setter := range setters {
		setter(opts)
	}
	return opts
}
