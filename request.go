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
	"net/url"
	"strings"

	"errors"

	"github.com/Wenchy/requests/internal/auth"
	"github.com/Wenchy/requests/internal/auth/redirector"
)

// Request is a wrapper of http.Request.
type Request struct {
	*http.Request
	opts *httpOptions
}

// WithContext returns a shallow copy of r.Request with its context changed to ctx.
// The provided ctx must be non-nil.
func (r *Request) WithContext(ctx context.Context) *Request {
	r.Request = r.Request.WithContext(ctx)
	return r
}

// newRequest wraps NewRequestWithContext using context.Background.
func newRequest(method, urlStr string, opts *httpOptions) (*Request, error) {
	if len(opts.Params) != 0 {
		// check raw URL, should not contain character '?'
		if strings.Contains(urlStr, "?") {
			return nil, errors.New("params not nil, so raw URL should not contain character '?'")
		}
		queryValues := url.Values{}
		for k, v := range opts.Params {
			queryValues.Add(k, v)
		}
		queryString := queryValues.Encode()
		urlStr += "?" + queryString
	}
	r, err := http.NewRequest(method, urlStr, opts.Body)
	if err != nil {
		return nil, err
	}
	// fill request headers
	if opts.Headers != nil {
		for k, v := range opts.Headers {
			r.Header.Set(k, v)
		}
	}
	// auth
	if opts.AuthInfo != nil {
		// TODO(wenchy): some other auth types
		if opts.AuthInfo.Type == auth.BasicAuth {
			r.SetBasicAuth(opts.AuthInfo.Username, opts.AuthInfo.Password)
		}
	}
	return &Request{Request: r, opts: opts}, nil
}

// do sends an HTTP request and returns an HTTP response, following policy
// (such as redirects, cookies, auth) as configured on the client.
func do(method, url string, opts *httpOptions) (*Response, error) {
	req, err := newRequest(method, url, opts)
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
	transport := env.transport
	if opts.DisableKeepAlives {
		// If option DisableKeepAlives set as true, then clone a new transport
		// just for this one-off HTTP request.
		transport = env.transport.Clone()
		transport.DisableKeepAlives = true
	}
	client := &Client{
		Client: &http.Client{
			CheckRedirect: redirector.RedirectPolicyFunc,
			Timeout:       opts.Timeout,
			Transport:     transport,
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
func request(method, url string, opts *httpOptions) (*Response, error) {
	return do(method, url, opts)
}

// requestData sends an HTTP request to the specified URL, with raw string
// as the request body.
func requestData(method, url string, opts *httpOptions) (*Response, error) {
	var body *strings.Reader
	if opts.Data != nil {
		d := fmt.Sprintf("%v", opts.Data)
		body = strings.NewReader(d)
	}
	// TODO: judge content type
	// opts.Headers["Content-Type"] = "application/x-www-form-urlencoded"
	opts.Body = body
	return do(method, url, opts)
}

// requestForm sends an HTTP request to the specified URL, with form's keys and
// values URL-encoded as the request body.
func requestForm(method, urlStr string, opts *httpOptions) (*Response, error) {
	var body *strings.Reader
	if opts.Form != nil {
		formValues := url.Values{}
		for k, v := range opts.Form {
			formValues.Add(k, v)
		}
		body = strings.NewReader(formValues.Encode())
	}
	opts.Headers["Content-Type"] = "application/x-www-form-urlencoded"
	opts.Body = body
	return do(method, urlStr, opts)
}

// requestJSON sends an HTTP request, and encode request body as json.
func requestJSON(method, url string, opts *httpOptions) (*Response, error) {
	var body *bytes.Buffer
	if opts.JSON != nil {
		reqBytes, err := json.Marshal(opts.JSON)
		if err != nil {
			return nil, err
		}
		body = bytes.NewBuffer(reqBytes)
	}

	opts.Headers["Content-Type"] = "application/json"
	opts.Body = body
	return do(method, url, opts)
}

// requestFiles sends an uploading request for multiple multipart-encoded files.
func requestFiles(method, url string, opts *httpOptions) (*Response, error) {
	var body bytes.Buffer
	bodyWriter := multipart.NewWriter(&body)
	if opts.Files != nil {
		for field, fh := range opts.Files {
			fileWriter, err := bodyWriter.CreateFormFile(field, fh.Name())
			if err != nil {
				return nil, err
			}
			if _, err := io.Copy(fileWriter, fh); err != nil {
				return nil, err
			}
		}
	}

	opts.Headers["Content-Type"] = bodyWriter.FormDataContentType()
	opts.Body = &body
	// write EOF before sending
	if err := bodyWriter.Close(); err != nil {
		return nil, err
	}
	return do(method, url, opts)
}

type bodyType int

const (
	bodyTypeDefault = iota
	bodyTypeData
	bodyTypeForm
	bodyTypeJSON
	bodyTypeFiles
)

func inferBodyType(opts *httpOptions) bodyType {
	if opts.Data != nil {
		return bodyTypeData
	} else if opts.Form != nil {
		return bodyTypeForm
	} else if opts.JSON != nil {
		return bodyTypeJSON
	} else if opts.Files != nil {
		return bodyTypeFiles
	} else {
		return bodyTypeDefault
	}
}

type dispatcher func(method, url string, opts *httpOptions) (*Response, error)

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
	bodyType := inferBodyType(opts)
	return dispatchers[bodyType](method, url, opts)
}
