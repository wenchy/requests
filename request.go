// Package requests is an elegant and simple HTTP library for Go, built for human beings.
//
// It mimics the classic Python Requests library (https://requests.readthedocs.io/).
package requests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"reflect"

	"github.com/Wenchy/requests/internal/auth"
)

// Request wraps [http.Request].
type Request struct {
	*http.Request
	opts *Options
	body []byte // auto filled from Request.Body
}

// Bytes returns the HTTP request body as []byte.
func (r *Request) Bytes() []byte {
	return r.body
}

// Text returns the HTTP request body as a string.
func (r *Request) Text() string {
	return string(r.body)
}

// newRequest creates a new HTTP request.
func newRequest(ctx context.Context, method, url string, opts *Options, body []byte) (*Request, error) {
	r, err := http.NewRequestWithContext(ctx, method, url, opts.Body)
	if err != nil {
		return nil, err
	}
	// query parameters
	if len(opts.Params) != 0 {
		q := r.URL.Query()
		for key, values := range opts.Params {
			for _, value := range values {
				q.Add(key, value)
			}
		}
		r.URL.RawQuery = q.Encode()
	}
	// headers
	for key, values := range opts.Headers {
		for _, value := range values {
			r.Header.Add(key, value)
		}
	}
	// auth
	if opts.AuthInfo != nil {
		// TODO(wenchy): some other auth types
		if opts.AuthInfo.Type == auth.BasicAuth {
			r.SetBasicAuth(opts.AuthInfo.Username, opts.AuthInfo.Password)
		}
	}
	return &Request{Request: r, opts: opts, body: body}, nil
}

// request sends an HTTP request.
func request(c *Client, method, url string, opts *Options) (*Response, error) {
	// NOTE: reading the body into memory is costly for large bodies.
	body := bytes.NewBuffer(nil)
	if opts.Body != nil {
		_, err := io.Copy(body, opts.Body)
		if err != nil {
			return nil, err
		}
	}
	opts.Body = body
	return c.request(method, url, opts, body.Bytes())
}

// requestData sends an HTTP request with opts.Data as the body. It deduces
// Content-Type only when the caller has not set one.
func requestData(c *Client, method, url string, opts *Options) (*Response, error) {
	body := bytes.NewBuffer(nil)
	if opts.Data != nil {
		dataBytes, err := dataToBody(opts.Data)
		if err != nil {
			return nil, err
		}
		// Deduce Content-Type only when the caller has not set one.
		if opts.Headers.Get("Content-Type") == "" {
			opts.Headers.Set("Content-Type", deduceContentType(opts.Data, dataBytes))
		}
		if _, err = body.Write(dataBytes); err != nil {
			return nil, err
		}
	}
	opts.Body = body
	return c.request(method, url, opts, body.Bytes())
}

// requestForm sends an HTTP request with form values URL-encoded as the body.
func requestForm(c *Client, method, url string, opts *Options) (*Response, error) {
	body := bytes.NewBuffer(nil)
	if opts.Form != nil {
		d := opts.Form.Encode()
		_, err := body.WriteString(d)
		if err != nil {
			return nil, err
		}
	}
	opts.Headers.Set("Content-Type", formContentType)
	opts.Body = body
	return c.request(method, url, opts, body.Bytes())
}

// requestJSON sends an HTTP request with opts.JSON encoded as JSON in the body.
func requestJSON(c *Client, method, url string, opts *Options) (*Response, error) {
	body := bytes.NewBuffer(nil)
	if opts.JSON != nil {
		d, err := json.Marshal(opts.JSON)
		if err != nil {
			return nil, err
		}
		_, err = body.Write(d)
		if err != nil {
			return nil, err
		}
	}
	opts.Headers.Set("Content-Type", jsonContentType)
	opts.Body = body
	return c.request(method, url, opts, body.Bytes())
}

// requestFiles sends an HTTP request with files multipart-encoded in the body.
func requestFiles(c *Client, method, url string, opts *Options) (*Response, error) {
	body := bytes.NewBuffer(nil)
	bodyWriter := multipart.NewWriter(body)
	if opts.Files != nil {
		for field, f := range opts.Files {
			fileWriter, err := bodyWriter.CreateFormFile(field, f.Name())
			if err != nil {
				return nil, err
			}
			if _, err := io.Copy(fileWriter, f); err != nil {
				return nil, err
			}
		}
	}
	// write EOF before sending
	if err := bodyWriter.Close(); err != nil {
		return nil, err
	}
	opts.Headers.Set("Content-Type", bodyWriter.FormDataContentType())
	opts.Body = body
	return c.request(method, url, opts, body.Bytes())
}

type bodyType int

const (
	bodyTypeDefault = iota
	bodyTypeData
	bodyTypeForm
	bodyTypeJSON
	bodyTypeFiles
)

type dispatcher func(c *Client, method, url string, opts *Options) (*Response, error)

var dispatchers map[bodyType]dispatcher = map[bodyType]dispatcher{
	bodyTypeDefault: request,
	bodyTypeData:    requestData,
	bodyTypeForm:    requestForm,
	bodyTypeJSON:    requestJSON,
	bodyTypeFiles:   requestFiles,
}

var (
	plainTextType   = "text/plain; charset=utf-8"
	jsonContentType = "application/json"
	formContentType = "application/x-www-form-urlencoded"
)

// dataToBody encodes data into request body bytes. It is the shared body
// production used whether or not Content-Type is deduced.
func dataToBody(data any) ([]byte, error) {
	if reader, ok := data.(io.Reader); ok {
		return io.ReadAll(reader)
	}
	bodyValue := reflect.Indirect(reflect.ValueOf(data))
	// A typed nil pointer (e.g. (*MyStruct)(nil)) reaches here as a non-nil
	// interface but yields an invalid reflect.Value after Indirect. Guard
	// against it so bodyValue.Interface() below does not panic.
	if !bodyValue.IsValid() {
		return fmt.Appendf(nil, "%v", data), nil
	}
	switch bodyValue.Kind() {
	case reflect.Struct, reflect.Map, reflect.Slice:
		// Assert against the (possibly dereferenced) bodyValue so that *[]byte
		// is treated as raw bytes instead of being JSON/base64-marshaled.
		if body, ok := bodyValue.Interface().([]byte); ok {
			return body, nil
		}
		return json.Marshal(data)
	default:
		return fmt.Appendf(nil, "%v", bodyValue.Interface()), nil
	}
}

// deduceContentType deduces the Content-Type for data from its type and the
// already-encoded body. It is only called when the caller has not set one.
func deduceContentType(data any, body []byte) string {
	if _, ok := data.(io.Reader); ok {
		return http.DetectContentType(body)
	}
	bodyValue := reflect.Indirect(reflect.ValueOf(data))
	if !bodyValue.IsValid() {
		return plainTextType
	}
	switch bodyValue.Kind() {
	case reflect.Struct, reflect.Map, reflect.Slice:
		// []byte (and *[]byte, after Indirect) is detected as raw bytes.
		if _, ok := bodyValue.Interface().([]byte); ok {
			return http.DetectContentType(body)
		}
		return jsonContentType
	default:
		return plainTextType
	}
}

// deduceContentTypeAndBody deduces the Content-Type and body from data.
func deduceContentTypeAndBody(data any) (string, []byte, error) {
	body, err := dataToBody(data)
	if err != nil {
		return "", nil, err
	}
	return deduceContentType(data, body), body, nil
}
