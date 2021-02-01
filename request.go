// Package requests is an elegant and simple HTTP library for golang, built for human beings.
//
// This package mimics the implementation of the classic Python package Requests(https://requests.readthedocs.io/)
package requests

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

// request issues a http request.
func request(method, rawurl string, headers map[string]string, params map[string]string, body io.Reader) (*Response, error) {
	if params != nil && len(params) != 0 {
		// check raw url, should not contain character '?'
		if strings.Contains(rawurl, "?") {
			return nil, errors.Errorf("params not nil, so raw url should not contain character '?'")
		}
		queryValues := url.Values{}
		for k, v := range params {
			queryValues.Add(k, v)
		}
		queryString := queryValues.Encode()
		rawurl += "?" + queryString
	}

	req, err := http.NewRequest(method, rawurl, body)
	if err != nil {
		return nil, err
	}

	// fill http headers
	if headers != nil {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	client := &http.Client{}
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

// requestForm issues a http request to the specified URL, with form's keys and
// values URL-encoded as the request body.
func requestForm(method, rawurl string, headers map[string]string, params map[string]string, form map[string]string) (*Response, error) {
	var body *strings.Reader
	if form != nil {
		formValues := url.Values{}
		for k, v := range form {
			formValues.Add(k, v)
		}
		body = strings.NewReader(formValues.Encode())
	}

	headers["Content-Type"] = "application/x-www-form-urlencoded"
	r, err := request(method, rawurl, headers, params, body)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// requestJSON issues a http request, and encode request body as json.
func requestJSON(method, rawurl string, headers map[string]string, params map[string]string, req interface{}) (*Response, error) {
	var body *bytes.Buffer
	if req != nil {
		reqBytes, err := json.Marshal(req)
		if err != nil {
			return nil, err
		}
		body = bytes.NewBuffer(reqBytes)
	}

	headers["Content-Type"] = "application/json"
	r, err := request(method, rawurl, headers, params, body)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// Get issues a http GET request.
func Get(url string, setters ...Option) (*Response, error) {
	opts := parseOptions(setters...)
	return request(http.MethodGet, url, opts.Headers, opts.Params, nil)
}

// Post issues a http POST request.
func Post(url string, setters ...Option) (*Response, error) {
	opts := parseOptions(setters...)
	if opts.Form != nil {
		return requestForm(http.MethodPost, url, opts.Headers, opts.Params, opts.Form)
	} else if opts.JSON != nil {
		return requestJSON(http.MethodPost, url, opts.Headers, opts.Params, opts.JSON)
	} else {
		return request(http.MethodPost, url, opts.Headers, opts.Params, nil)
	}
}

// Put issues a http PUT request.
func Put(url string, setters ...Option) (*Response, error) {
	opts := parseOptions(setters...)
	if opts.Form != nil {
		return requestForm(http.MethodPut, url, opts.Headers, opts.Params, opts.Form)
	} else if opts.JSON != nil {
		return requestJSON(http.MethodPut, url, opts.Headers, opts.Params, opts.JSON)
	} else {
		return request(http.MethodPut, url, opts.Headers, opts.Params, nil)
	}
}

// Delete issues a http DELETE request.
func Delete(url string, setters ...Option) (*Response, error) {
	opts := parseOptions(setters...)
	return request(http.MethodDelete, url, opts.Headers, opts.Params, nil)
}
