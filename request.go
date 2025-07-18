// Package requests is an elegant and simple HTTP library for golang, built for human beings.
//
// This package mimics the implementation of the classic Python package Requests(https://requests.readthedocs.io/)
package requests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

// Request is a wrapper of http.Request.
type Request struct {
	*http.Request
	opts *Options
	body []byte // auto filled from Request.Body
}

// WithContext returns a shallow copy of r.Request with its context changed to ctx.
// The provided ctx must be non-nil.
func (r *Request) WithContext(ctx context.Context) {
	r.Request = r.Request.WithContext(ctx)
}

// Bytes returns the HTTP request body as []byte.
func (r *Request) Bytes() []byte {
	return r.body
}

// Text parses the HTTP request body as string.
func (r *Request) Text() string {
	return string(r.body)
}

// newRequest creates a new HTTP request.
func newRequest(method, url string, opts *Options, body []byte) (*Request, error) {
	r, err := http.NewRequest(method, url, opts.Body)
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
		if opts.AuthInfo.Type == AuthTypeBasic {
			r.SetBasicAuth(opts.AuthInfo.Username, opts.AuthInfo.Password)
		}
	}
	return &Request{Request: r, opts: opts, body: body}, nil
}

// request sends an HTTP request.
func request(c *Client, method, url string, opts *Options) (*Response, error) {
	// NOTE: get the body size from io.Reader. It is costy for large body.
	body := bytes.NewBuffer(nil)
	if opts.Body != nil {
		_, err := io.Copy(body, opts.Body)
		if err != nil {
			return nil, err
		}
	}
	opts.Body = body
	return c.Do(method, url, opts, body.Bytes())
}

// requestData sends an HTTP request to the specified URL, with raw string
// as the request body.
func requestData(c *Client, method, url string, opts *Options) (*Response, error) {
	body := bytes.NewBuffer(nil)
	if opts.Data != nil {
		d := fmt.Sprintf("%v", opts.Data)
		_, err := body.WriteString(d)
		if err != nil {
			return nil, err
		}
	}
	// TODO: judge content type
	// opts.Headers["Content-Type"] = "application/x-www-form-urlencoded"
	opts.Body = body
	return c.Do(method, url, opts, body.Bytes())
}

// requestForm sends an HTTP request to the specified URL, with form's keys and
// values URL-encoded as the request body.
func requestForm(c *Client, method, url string, opts *Options) (*Response, error) {
	body := bytes.NewBuffer(nil)
	if opts.Form != nil {
		d := opts.Form.Encode()
		_, err := body.WriteString(d)
		if err != nil {
			return nil, err
		}
	}
	opts.Headers.Set("Content-Type", "application/x-www-form-urlencoded")
	opts.Body = body
	return c.Do(method, url, opts, body.Bytes())
}

// requestJSON sends an HTTP request, and encode request body as json.
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
	opts.Headers.Set("Content-Type", "application/json")
	opts.Body = body
	return c.Do(method, url, opts, body.Bytes())
}

// requestFiles sends an uploading request for multiple multipart-encoded files.
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
	return c.Do(method, url, opts, body.Bytes())
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

func (c *Client) callMethod(method, url string, options ...Option) (*Response, error) {
	opts := parseOptions(options...)
	return dispatchers[opts.bodyType](c, method, url, opts)
}
