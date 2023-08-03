// Package requests is an elegant and simple HTTP library for golang, built for human beings.
//
// This package mimics the implementation of the classic Python package Requests(https://requests.readthedocs.io/)
package requests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	"errors"
)

// request sends an HTTP request.
func request(method, urlStr string, options ...Option) (*Response, error) {
	opts := parseOptions(options...)
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

	req, err := http.NewRequest(method, urlStr, opts.Body)
	if err != nil {
		return nil, err
	}

	// fill request headers
	if opts.Headers != nil {
		for k, v := range opts.Headers {
			req.Header.Set(k, v)
		}
	}

	// TODO(wenchy): some other auth types
	if opts.Auth.authType == HTTPBasicAuth {
		req.SetBasicAuth(opts.Auth.username, opts.Auth.password)
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
	transport := env.DefaultTransport
	if opts.DisableKeepAlives {
		// If option DisableKeepAlives set as true, then clone a new transport
		// just for this one-off HTTP request.
		transport = env.DefaultTransport.Clone()
		transport.DisableKeepAlives = true
	}
	client := &http.Client{
		CheckRedirect: redirectPolicyFunc,
		Timeout:       opts.Timeout,
		Transport:     transport,
	}

	// If the returned error is nil, the Response will contain
	// a non-nil Body which the user is expected to close.
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	// wrap response
	r := &Response{
		resp: resp,
	}
	// return error with status and text body embedded if status code
	// is not 2XX, and response is also returned.
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		txt, _ := r.Text()
		return r, errors.New(resp.Status + " " + txt)
	}

	return r, nil
}

// requestData sends an HTTP request to the specified URL, with raw string
// as the request body.
func requestData(method, urlStr string, options ...Option) (*Response, error) {
	opts := parseOptions(options...)
	var body *strings.Reader
	if opts.Data != nil {
		d := fmt.Sprintf("%v", opts.Data)
		body = strings.NewReader(d)
	}
	// TODO: judge content type
	// opts.Headers["Content-Type"] = "application/x-www-form-urlencoded"

	// options = append(options, Headers(opts.Headers))
	options = append(options, Body(body))
	r, err := request(method, urlStr, options...)
	if err != nil {
		return r, err
	}

	return r, nil
}

// requestForm sends an HTTP request to the specified URL, with form's keys and
// values URL-encoded as the request body.
func requestForm(method, urlStr string, options ...Option) (*Response, error) {
	opts := parseOptions(options...)
	var body *strings.Reader
	if opts.Form != nil {
		formValues := url.Values{}
		for k, v := range opts.Form {
			formValues.Add(k, v)
		}
		body = strings.NewReader(formValues.Encode())
	}
	opts.Headers["Content-Type"] = "application/x-www-form-urlencoded"

	options = append(options, Headers(opts.Headers))
	options = append(options, Body(body))
	r, err := request(method, urlStr, options...)
	if err != nil {
		return r, err
	}

	return r, nil
}

// requestJSON sends an HTTP request, and encode request body as json.
func requestJSON(method, url string, options ...Option) (*Response, error) {
	opts := parseOptions(options...)
	var body *bytes.Buffer
	if opts.JSON != nil {
		reqBytes, err := json.Marshal(opts.JSON)
		if err != nil {
			return nil, err
		}
		body = bytes.NewBuffer(reqBytes)
	}

	opts.Headers["Content-Type"] = "application/json"

	options = append(options, Headers(opts.Headers))
	options = append(options, Body(body))
	r, err := request(method, url, options...)
	if err != nil {
		return r, err
	}

	return r, nil
}

// requestFiles sends an uploading request for multiple multipart-encoded files.
func requestFiles(method, url string, options ...Option) (*Response, error) {
	opts := parseOptions(options...)
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
	
	options = append(options, Headers(opts.Headers))
	options = append(options, Body(&body))
	// write EOF before sending
	if err := bodyWriter.Close(); err != nil {
		return nil, err
	}
	return request(method, url, options...)
}

// Get sends an HTTP GET request.
func Get(url string, options ...Option) (*Response, error) {
	return request(http.MethodGet, url, options...)
}

// Post sends an HTTP POST request.
func Post(url string, options ...Option) (*Response, error) {
	opts := parseOptions(options...)
	if opts.Data != nil {
		return requestData(http.MethodPost, url, options...)
	} else if opts.Form != nil {
		return requestForm(http.MethodPost, url, options...)
	} else if opts.JSON != nil {
		return requestJSON(http.MethodPost, url, options...)
	} else if opts.Files != nil {
		return requestFiles(http.MethodPost, url, options...)
	} else {
		return request(http.MethodPost, url, options...)
	}
}

// Put sends an HTTP PUT request.
func Put(url string, options ...Option) (*Response, error) {
	opts := parseOptions(options...)
	if opts.Data != nil {
		return requestData(http.MethodPut, url, options...)
	} else if opts.Form != nil {
		return requestForm(http.MethodPut, url, options...)
	} else if opts.JSON != nil {
		return requestJSON(http.MethodPut, url, options...)
	} else {
		return request(http.MethodPut, url, options...)
	}
}

// Patch sends an HTTP PATCH request.
func Patch(url string, options ...Option) (*Response, error) {
	opts := parseOptions(options...)
	if opts.Data != nil {
		return requestData(http.MethodPatch, url, options...)
	} else if opts.Form != nil {
		return requestForm(http.MethodPatch, url, options...)
	} else if opts.JSON != nil {
		return requestJSON(http.MethodPatch, url, options...)
	} else {
		return request(http.MethodPatch, url, options...)
	}
}

// Delete sends an HTTP DELETE request.
func Delete(url string, options ...Option) (*Response, error) {
	opts := parseOptions(options...)
	if opts.Data != nil {
		return requestData(http.MethodDelete, url, options...)
	} else if opts.Form != nil {
		return requestForm(http.MethodDelete, url, options...)
	} else if opts.JSON != nil {
		return requestJSON(http.MethodDelete, url, options...)
	} else {
		return request(http.MethodDelete, url, options...)
	}
}
