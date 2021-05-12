// Package requests is an elegant and simple HTTP library for golang, built for human beings.
//
// This package mimics the implementation of the classic Python package Requests(https://requests.readthedocs.io/)
package requests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// request issues a http request.
func request(method, rawurl string, setters ...Option) (*Response, error) {
	opts := parseOptions(setters...)
	if opts.Params != nil && len(opts.Params) != 0 {
		// check raw url, should not contain character '?'
		if strings.Contains(rawurl, "?") {
			return nil, errors.Errorf("params not nil, so raw url should not contain character '?'")
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

	// fill http headers
	if opts.Headers != nil {
		for k, v := range opts.Headers {
			req.Header.Set(k, v)
		}
	}

	if opts.Auth.authType == HTTPBasicAuth {
		req.SetBasicAuth(opts.Auth.username, opts.Auth.password)
	}
	// TODO(wenchy): some other auth types

	client := &http.Client{
		CheckRedirect: redirectPolicyFunc,
		Timeout:       time.Duration(opts.Timeout) * time.Second,
	}
	// fmt.Printf("timeout: %d\n", opts.Timeout)
	rsp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	// wrap http response
	r := &Response{
		rsp: rsp,
	}

	if rsp.StatusCode != http.StatusOK {
		return r, errors.Errorf(rsp.Status)
	}

	return r, nil
}

// requestData issues a http request to the specified URL, with raw string
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

// requestForm issues a http request to the specified URL, with form's keys and
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

// requestJSON issues a http request, and encode request body as json.
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

// Get issues a http GET request.
func Get(rawurl string, setters ...Option) (*Response, error) {
	return request(http.MethodGet, rawurl, setters...)
}

// Post issues a http POST request.
func Post(rawurl string, setters ...Option) (*Response, error) {
	opts := parseOptions(setters...)
	if opts.Data != nil {
		return requestData(http.MethodPost, rawurl, setters...)
	} else if opts.Form != nil {
		return requestForm(http.MethodPost, rawurl, setters...)
	} else if opts.JSON != nil {
		return requestJSON(http.MethodPost, rawurl, setters...)
	} else {
		return request(http.MethodPost, rawurl, setters...)
	}
}

// Put issues a http PUT request.
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

// Delete issues a http DELETE request.
func Delete(rawurl string, setters ...Option) (*Response, error) {
	return request(http.MethodDelete, rawurl, setters...)
}
