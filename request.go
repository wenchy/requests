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

// requestData sends an HTTP request with opts.Data as the body.
func requestData(c *Client, method, url string, opts *Options) (*Response, error) {
	body := bytes.NewBuffer(nil)
	if opts.Data != nil {
		contentType, bytes, err := deduceContentTypeAndBody(opts.Data)
		if err != nil {
			return nil, err
		}
		_, err = body.Write(bytes)
		if err != nil {
			return nil, err
		}
		opts.Headers.Set("Content-Type", contentType)
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

// deduceContentTypeAndBody deduces the Content-Type and body from data.
func deduceContentTypeAndBody(data any) (string, []byte, error) {
	if reader, ok := data.(io.Reader); ok {
		body, err := io.ReadAll(reader)
		return http.DetectContentType(body), body, err
	}
	bodyValue := reflect.Indirect(reflect.ValueOf(data))
	// A typed nil pointer (e.g. (*MyStruct)(nil)) reaches here as a non-nil
	// interface but yields an invalid reflect.Value after Indirect. Guard
	// against it so bodyValue.Interface() below does not panic.
	if !bodyValue.IsValid() {
		return plainTextType, fmt.Appendf(nil, "%v", data), nil
	}
	switch bodyValue.Kind() {
	case reflect.Struct, reflect.Map, reflect.Slice:
		// check slice here to differentiate between any slice vs byte slice.
		// Assert against the (possibly dereferenced) bodyValue so that *[]byte
		// is treated as raw bytes instead of being JSON/base64-marshaled.
		if body, ok := bodyValue.Interface().([]byte); ok {
			return http.DetectContentType(body), body, nil
		}
		body, err := json.Marshal(data)
		return jsonContentType, body, err
	default:
		return plainTextType, fmt.Appendf(nil, "%v", bodyValue.Interface()), nil
	}
}
