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
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"errors"
)

// request sends an HTTP request.
func request(method, rawurl string, setters ...Option) (*Response, error) {
	opts := parseOptions(setters...)
	if len(opts.Params) != 0 {
		// check raw url, should not contain character '?'
		if strings.Contains(rawurl, "?") {
			return nil, errors.New("params not nil, so raw url should not contain character '?'")
		}
		queryValues := url.Values{}
		for k, v := range opts.Params {
			queryValues.Add(k, v)
		}
		queryString := queryValues.Encode()
		rawurl += "?" + queryString
	}

	req, err := http.NewRequest(method, rawurl, opts.Body)
	if err != nil {
		return nil, err
	}

	// fill request headers
	if opts.Headers != nil {
		for k, v := range opts.Headers {
			req.Header.Set(k, v)
		}
	}

	if opts.Auth.authType == HTTPBasicAuth {
		req.SetBasicAuth(opts.Auth.username, opts.Auth.password)
	}
	// TODO(wenchy): some other auth types
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DisableKeepAlives:     opts.DisableKeepAlives,
	}
	client := &http.Client{
		CheckRedirect: redirectPolicyFunc,
		Timeout:       time.Duration(opts.Timeout) * time.Second,
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
func requestData(method, rawurl string, setters ...Option) (*Response, error) {
	opts := parseOptions(setters...)
	var body *strings.Reader
	if opts.Data != nil {
		d := fmt.Sprintf("%v", opts.Data)
		body = strings.NewReader(d)
	}
	// TODO: judge content type
	// opts.Headers["Content-Type"] = "application/x-www-form-urlencoded"

	// setters = append(setters, Headers(opts.Headers))
	setters = append(setters, Body(body))
	r, err := request(method, rawurl, setters...)
	if err != nil {
		return r, err
	}

	return r, nil
}

// requestForm sends an HTTP request to the specified URL, with form's keys and
// values URL-encoded as the request body.
func requestForm(method, rawurl string, setters ...Option) (*Response, error) {
	opts := parseOptions(setters...)
	var body *strings.Reader
	if opts.Form != nil {
		formValues := url.Values{}
		for k, v := range opts.Form {
			formValues.Add(k, v)
		}
		body = strings.NewReader(formValues.Encode())
	}
	opts.Headers["Content-Type"] = "application/x-www-form-urlencoded"

	setters = append(setters, Headers(opts.Headers))
	setters = append(setters, Body(body))
	r, err := request(method, rawurl, setters...)
	if err != nil {
		return r, err
	}

	return r, nil
}

// requestJSON sends an HTTP request, and encode request body as json.
func requestJSON(method, rawurl string, setters ...Option) (*Response, error) {
	opts := parseOptions(setters...)
	var body *bytes.Buffer
	if opts.JSON != nil {
		reqBytes, err := json.Marshal(opts.JSON)
		if err != nil {
			return nil, err
		}
		body = bytes.NewBuffer(reqBytes)
	}

	opts.Headers["Content-Type"] = "application/json"

	setters = append(setters, Headers(opts.Headers))
	setters = append(setters, Body(body))
	r, err := request(method, rawurl, setters...)
	if err != nil {
		return r, err
	}

	return r, nil
}

// requestFiles sends an uploading request for multiple multipart-encoded files.
func requestFiles(method, rawurl string, setters ...Option) (*Response, error) {
	opts := parseOptions(setters...)
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

	setters = append(setters, Headers(opts.Headers))
	setters = append(setters, Body(&body))
	// write EOF before sending
	if err := bodyWriter.Close(); err != nil {
		return nil, err
	}
	return request(method, rawurl, setters...)
}

// Get sends an HTTP GET request.
func Get(rawurl string, setters ...Option) (*Response, error) {
	return request(http.MethodGet, rawurl, setters...)
}

// Post sends an HTTP POST request.
func Post(rawurl string, setters ...Option) (*Response, error) {
	opts := parseOptions(setters...)
	if opts.Data != nil {
		return requestData(http.MethodPost, rawurl, setters...)
	} else if opts.Form != nil {
		return requestForm(http.MethodPost, rawurl, setters...)
	} else if opts.JSON != nil {
		return requestJSON(http.MethodPost, rawurl, setters...)
	} else if opts.Files != nil {
		return requestFiles(http.MethodPost, rawurl, setters...)
	} else {
		return request(http.MethodPost, rawurl, setters...)
	}
}

// Put sends an HTTP PUT request.
func Put(rawurl string, setters ...Option) (*Response, error) {
	opts := parseOptions(setters...)
	if opts.Data != nil {
		return requestData(http.MethodPut, rawurl, setters...)
	} else if opts.Form != nil {
		return requestForm(http.MethodPut, rawurl, setters...)
	} else if opts.JSON != nil {
		return requestJSON(http.MethodPut, rawurl, setters...)
	} else {
		return request(http.MethodPut, rawurl, setters...)
	}
}

// Patch sends an HTTP PATCH request.
func Patch(rawurl string, setters ...Option) (*Response, error) {
	opts := parseOptions(setters...)
	if opts.Data != nil {
		return requestData(http.MethodPatch, rawurl, setters...)
	} else if opts.Form != nil {
		return requestForm(http.MethodPatch, rawurl, setters...)
	} else if opts.JSON != nil {
		return requestJSON(http.MethodPatch, rawurl, setters...)
	} else {
		return request(http.MethodPatch, rawurl, setters...)
	}
}

// Delete sends an HTTP DELETE request.
func Delete(rawurl string, setters ...Option) (*Response, error) {
	opts := parseOptions(setters...)
	if opts.Data != nil {
		return requestData(http.MethodDelete, rawurl, setters...)
	} else if opts.Form != nil {
		return requestForm(http.MethodDelete, rawurl, setters...)
	} else if opts.JSON != nil {
		return requestJSON(http.MethodDelete, rawurl, setters...)
	} else {
		return request(http.MethodDelete, rawurl, setters...)
	}
}
