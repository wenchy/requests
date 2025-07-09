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

	"github.com/Wenchy/requests/internal/auth"
	"github.com/Wenchy/requests/internal/auth/redirector"
)

// Request is a wrapper of http.Request.
type Request struct {
	*http.Request
	opts *Options
	body []byte // auto filled from Request.Body
}

// WithContext returns a shallow copy of r.Request with its context changed to ctx.
// The provided ctx must be non-nil.
func (r *Request) WithContext(ctx context.Context) *Request {
	r.Request = r.Request.WithContext(ctx)
	return r
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
		if opts.AuthInfo.Type == auth.BasicAuth {
			r.SetBasicAuth(opts.AuthInfo.Username, opts.AuthInfo.Password)
		}
	}
	return &Request{Request: r, opts: opts, body: body}, nil
}

// do sends an HTTP request and returns an HTTP response, following policy
// (such as redirects, cookies, auth) as configured on the client.
func do(method, url string, opts *Options, body []byte) (*Response, error) {
	req, err := newRequest(method, url, opts, body)
	if err != nil {
		return nil, err
	}

	// NOTE: Keep-Alive & Connection Pooling
	//
	// 1. Keep-Alive
	//
	// 	The net/http Transport documentation uses the term to refer to
	//  persistent connections. A keep-alive or persistent connection
	//  is a connection that can be used for more than one HTTP
	//  transaction.
	//
	// 	The Transport.IdleConnTimeout field specifies how long the
	//  transport keeps an unused connection in the pool before closing
	//  the connection.
	//
	//  The net Dialer documentation uses the keep-alive term to refer
	//  the TCP feature for probing the health of a connection.
	// 	Dialer.KeepAlive field specifies how frequently TCP keep-alive
	//  probes are sent to the peer.
	//
	// 2. Connection Pooling
	//
	//  Connections are added to the pool in the function
	//  Transport.tryPutIdleConn. The connection is not pooled if
	//  Transport.DisableKeepAlives is true or Transport.MaxIdleConnsPerHost
	//  is less than zero.
	//
	//  Setting either value disables pooling. The transport adds the
	//  "Connection: close" request header when DisableKeepAlives is true.
	//  This may or may not be desirable depending on what you are testing.
	//
	// 3. References:
	//
	// - https://stackoverflow.com/questions/57683132/turning-off-connection-pool-for-go-http-client
	// - https://stackoverflow.com/questions/59656164/what-is-the-difference-between-net-dialerkeepalive-and-http-transportidletimeo
	var roundTripper http.RoundTripper
	if opts.RoundTripper != nil {
		roundTripper = opts.RoundTripper
	} else {
		if rt := env.hostRoundTrippers[req.Host]; rt != nil {
			// Use the host-specific RoundTripper if set.
			roundTripper = rt
		} else if opts.DisableKeepAlives {
			// If option DisableKeepAlives set as true, then clone a new transport
			// just for this one-off HTTP request.
			transport := env.transport.Clone()
			transport.DisableKeepAlives = true
			roundTripper = transport
		} else {
			// If option DisableKeepAlives not set as true, then use the default
			// transport.
			roundTripper = env.transport
		}
	}
	client := &Client{
		Client: &http.Client{
			CheckRedirect: redirector.RedirectPolicyFunc,
			Timeout:       opts.Timeout,
			Transport:     roundTripper,
		},
	}
	var ctx context.Context
	if opts.ctx != nil {
		ctx = opts.ctx // use ctx from options if set
	} else {
		newCtx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
		defer cancel()
		ctx = newCtx
	}
	return client.Do(ctx, req)
}

// request sends an HTTP request.
func request(method, url string, opts *Options) (*Response, error) {
	// NOTE: get the body size from io.Reader. It is costy for large body.
	body := bytes.NewBuffer(nil)
	if opts.Body != nil {
		_, err := io.Copy(body, opts.Body)
		if err != nil {
			return nil, err
		}
	}
	opts.Body = body
	return do(method, url, opts, body.Bytes())
}

// requestData sends an HTTP request to the specified URL, with raw string
// as the request body.
func requestData(method, url string, opts *Options) (*Response, error) {
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
	return do(method, url, opts, body.Bytes())
}

// requestForm sends an HTTP request to the specified URL, with form's keys and
// values URL-encoded as the request body.
func requestForm(method, url string, opts *Options) (*Response, error) {
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
	return do(method, url, opts, body.Bytes())
}

// requestJSON sends an HTTP request, and encode request body as json.
func requestJSON(method, url string, opts *Options) (*Response, error) {
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
	return do(method, url, opts, body.Bytes())
}

// requestFiles sends an uploading request for multiple multipart-encoded files.
func requestFiles(method, url string, opts *Options) (*Response, error) {
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
	return do(method, url, opts, body.Bytes())
}

type bodyType int

const (
	bodyTypeDefault = iota
	bodyTypeData
	bodyTypeForm
	bodyTypeJSON
	bodyTypeFiles
)

type dispatcher func(method, url string, opts *Options) (*Response, error)

var dispatchers map[bodyType]dispatcher

func init() {
	dispatchers = map[bodyType]dispatcher{
		bodyTypeDefault: request,
		bodyTypeData:    requestData,
		bodyTypeForm:    requestForm,
		bodyTypeJSON:    requestJSON,
		bodyTypeFiles:   requestFiles,
	}
}

func callMethod(method, url string, options ...Option) (*Response, error) {
	opts := parseOptions(options...)
	return dispatchers[opts.bodyType](method, url, opts)
}
