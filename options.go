package requests

import (
	"fmt"
)

// Options follow the design of Functional Options(https://github.com/tmrts/go-patterns/blob/master/idiom/functional-options.md)
type Options struct {
	Headers map[string]string
	Params  map[string]string
	Data    interface{}
	Form    map[string]string
	JSON    interface{}
}

// Option is the functional option type.
type Option func(*Options)

// Headers set the HTTP header KVs.
func Headers(headers map[string]string) Option {
	return func(opts *Options) {
		opts.Headers = headers
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
		opts.Params = params
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

func newDefaultOptions() *Options {
	return &Options{
		Headers: map[string]string{},
		Params:  map[string]string{},
		Form:    nil,
		JSON:    nil,
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
