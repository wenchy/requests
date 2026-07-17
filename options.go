package requests

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/Wenchy/requests/internal/auth"
)

// Options defines the optional parameters for an HTTP request.
type Options struct {
	ctx context.Context

	Headers http.Header
	Params  url.Values

	// body
	bodyType bodyType
	Body     io.Reader
	// different request body types
	Data  any
	Form  url.Values
	JSON  any
	Files map[string]*os.File

	// different response body types
	ToText *string
	ToJSON any

	// auth
	AuthInfo *auth.AuthInfo
	// request timeout
	Timeout time.Duration

	// dump
	DumpRequestOut *string
	DumpResponse   *string

	// interceptor
	Interceptor InterceptorFunc
}

// Option is the functional option type.
type Option func(*Options)

// newDefaultOptions creates default HTTP options.
func newDefaultOptions() *Options {
	return &Options{
		ctx:      context.Background(),
		Headers:  http.Header{},
		bodyType: bodyTypeDefault,
	}
}

func parseOptions(options ...Option) *Options {
	opts := newDefaultOptions()
	for _, setter := range options {
		setter(opts)
	}
	return opts
}

// Context sets the HTTP request context.
//
// For an outgoing client request, the context controls the entire lifetime of
// a request and its response: obtaining a connection, sending the request,
// and reading the response headers and body.
func Context(ctx context.Context) Option {
	return func(opts *Options) {
		opts.ctx = ctx
	}
}

// Headers sets the HTTP headers. The keys should be in canonical form, as
// returned by [http.CanonicalHeaderKey]. Two types are supported:
//
// # Type 1: map[string]string
//
//	map[string]string(
//		"Header-Key1", "val1",
//		"Header-Key2", "val2",
//	)
//
// # Type 2: http.Header
//
//	http.Header(
//		"Header-Key1", []string{"val1", "val1-2"},
//		"Header-Key2", "val2",
//	)
//
// [http.CanonicalHeaderKey]: https://pkg.go.dev/net/http#CanonicalHeaderKey
func Headers[T map[string]string | http.Header](headers T) Option {
	return func(opts *Options) {
		switch headers := any(headers).(type) {
		case map[string]string:
			for k, v := range headers {
				opts.Headers.Add(k, v)
			}
		case http.Header:
			for key, values := range headers {
				for _, value := range values {
					opts.Headers.Add(key, value)
				}
			}
		}
	}
}

// HeaderPairs sets HTTP headers from key-value pairs.
// The keys should be in canonical form, as returned by
// [http.CanonicalHeaderKey]. It panics if len(kv) is odd.
//
// Values with the same key will be merged into a list:
//
//	HeaderPairs(
//		"Key1", "val1",
//		"Key1", "val1-2", // "Key1" will have map value []string{"val1", "val1-2"}
//		"Key2", "val2",
//	)
//
// [http.CanonicalHeaderKey]: https://pkg.go.dev/net/http#CanonicalHeaderKey
func HeaderPairs(kv ...string) Option {
	if len(kv)%2 == 1 {
		panic(fmt.Sprintf("params: got the odd number of input pairs: %d", len(kv)))
	}
	headers := http.Header{}
	var key string
	for i, s := range kv {
		if i%2 == 0 {
			key = s
			continue
		}
		headers.Add(key, s)
	}
	return Headers(headers)
}

// Params sets the URL query string parameters. Two types are supported:
//
// # Type 1: map[string]string
//
//	map[string]string(
//		"key1", "val1",
//		"key2", "val2",
//	)
//
// # Type 2: url.Values
//
//	url.Values(
//		"key1", []string{"val1", "val1-2"},
//		"key2", "val2",
//	)
func Params[T map[string]string | url.Values](params T) Option {
	return func(opts *Options) {
		if opts.Params == nil {
			opts.Params = url.Values{}
		}
		switch params := any(params).(type) {
		case map[string]string:
			for k, v := range params {
				opts.Params.Add(k, v)
			}
		case url.Values:
			for key, values := range params {
				for _, value := range values {
					opts.Params.Add(key, value)
				}
			}
		}
	}
}

// ParamPairs sets query parameters from key-value pairs.
// It panics if len(kv) is odd.
//
// Values with the same key will be merged into a list:
//
//	ParamPairs(
//		"key1", "val1",
//		"key1", "val1-2", // "key1" will have map value []string{"val1", "val1-2"}
//		"key2", "val2",
//	)
func ParamPairs(kv ...string) Option {
	if len(kv)%2 == 1 {
		panic(fmt.Sprintf("params: got the odd number of input pairs: %d", len(kv)))
	}
	params := url.Values{}
	var key string
	for i, s := range kv {
		if i%2 == 0 {
			key = s
			continue
		}
		params.Add(key, s)
	}
	return Params(params)
}

// Body sets the request body from an [io.Reader].
func Body(body io.Reader) Option {
	return func(opts *Options) {
		opts.Body = body
		opts.bodyType = bodyTypeDefault
	}
}

// Data sets the request body from data, deducing Content-Type from its type:
//
//   - io.Reader (e.g. *bytes.Buffer): read and detected via [http.DetectContentType]
//   - []byte or *[]byte: detected via [http.DetectContentType]
//   - struct, map, slice (except []byte): "application/json"
//   - otherwise: "text/plain", formatted with %v
//
// An explicit Content-Type set via [Headers] or [HeaderPairs] takes
// precedence over the deduced value. Non-reader pointers are dereferenced
// before the kind is determined, so *struct and *[]byte behave like their
// pointed-to values. A typed nil pointer such as (*T)(nil) falls into the
// last case and is sent as "<nil>".
func Data(data any) Option {
	return func(opts *Options) {
		opts.Data = data
		opts.bodyType = bodyTypeData
	}
}

// Form sets the request body from form values, and forces Content-Type to
// "application/x-www-form-urlencoded" (overriding any caller value). Two
// types are supported:
//
// # Type 1: map[string]string
//
//	map[string]string(
//		"key1", "val1",
//		"key2", "val2",
//	)
//
// # Type 2: url.Values
//
//	url.Values(
//		"key1", []string{"val1", "val1-2"},
//		"key2", "val2",
//	)
func Form[T map[string]string | url.Values](params T) Option {
	return func(opts *Options) {
		if opts.Form == nil {
			opts.Form = url.Values{}
		}
		switch params := any(params).(type) {
		case map[string]string:
			for k, v := range params {
				opts.Form.Add(k, v)
			}
		case url.Values:
			for key, values := range params {
				for _, value := range values {
					opts.Form.Add(key, value)
				}
			}
		}
		opts.bodyType = bodyTypeForm
	}
}

// FormPairs sets form values from key-value pairs.
// It panics if len(kv) is odd.
//
// Values with the same key will be merged into a list:
//
//	FormPairs(
//		"key1", "val1",
//		"key1", "val1-2", // "key1" will have map value []string{"val1", "val1-2"}
//		"key2", "val2",
//	)
func FormPairs(kv ...string) Option {
	if len(kv)%2 == 1 {
		panic(fmt.Sprintf("params: got the odd number of input pairs: %d", len(kv)))
	}
	form := url.Values{}
	var key string
	for i, s := range kv {
		if i%2 == 0 {
			key = s
			continue
		}
		form.Add(key, s)
	}
	return Form(form)
}

// JSON marshals v as JSON into the request body. It sets Content-Type to
// "application/json" unless the caller has already set one.
func JSON(v any) Option {
	return func(opts *Options) {
		opts.JSON = v
		opts.bodyType = bodyTypeJSON
	}
}

// Files sets files as a map of field to file handler, and forces
// Content-Type to "multipart/form-data" with the multipart boundary
// (overriding any caller value).
func Files(files map[string]*os.File) Option {
	return func(opts *Options) {
		if opts.Files != nil {
			for k, v := range files {
				opts.Files[k] = v
			}
		} else {
			opts.Files = files
		}
		opts.bodyType = bodyTypeFiles
	}
}

// ToText captures the response body as text into the string pointed to by v.
func ToText(v *string) Option {
	return func(opts *Options) {
		opts.ToText = v
	}
}

// ToJSON decodes the response body into v as JSON.
func ToJSON(v any) Option {
	return func(opts *Options) {
		opts.ToJSON = v
	}
}

// BasicAuth enables HTTP Basic Auth with the given username and password.
func BasicAuth(username, password string) Option {
	return func(opts *Options) {
		opts.AuthInfo = &auth.AuthInfo{
			Type:     auth.BasicAuth,
			Username: username,
			Password: password,
		}
	}
}

// Timeout sets the HTTP request context timeout by [context.WithTimeout].
func Timeout(timeout time.Duration) Option {
	return func(opts *Options) {
		opts.Timeout = timeout
	}
}

// Dump dumps the outgoing request into req and the response into resp,
// when the corresponding pointer is non-nil.
//
// Refer:
//   - https://pkg.go.dev/net/http/httputil#DumpRequestOut
//   - https://pkg.go.dev/net/http/httputil#DumpResponse
func Dump(req, resp *string) Option {
	return func(opts *Options) {
		opts.DumpRequestOut = req
		opts.DumpResponse = resp
	}
}

// Interceptor prepends an interceptor to the client interceptors for the
// current request only.
func Interceptor(interceptor InterceptorFunc) Option {
	return func(opts *Options) {
		opts.Interceptor = interceptor
	}
}
