package requests

import (
	"encoding/json"
	"io"
	"net/http"
)

// Response is a wrapper of HTTP response.
type Response struct {
	resp        *http.Response
	body       []byte // filled if resp.Body was already drained.
	bodyClosed bool
}

// StatusCode returns status code of HTTP response.
func (r *Response) StatusCode() int {
	if r.resp == nil {
		return http.StatusServiceUnavailable
	}
	return r.resp.StatusCode
}

// Raw returns raw body of response.
func (r *Response) Raw() io.ReadCloser {
	return r.resp.Body
}

// Bytes parses the HTTP response body as []byte.
func (r *Response) Bytes() ([]byte, error) {
	if !r.bodyClosed {
		if err := r.readAll(); err != nil {
			return nil, err
		}
	}
	return r.body, nil
}

// Text parses the HTTP response body as string.
func (r *Response) Text() (string, error) {
	if !r.bodyClosed {
		if err := r.readAll(); err != nil {
			return "", err
		}
	}
	return string(r.body), nil
}

// JSON decodes the HTTP response body as JSON format.
func (r *Response) JSON(v interface{}) error {
	if !r.bodyClosed {
		if err := r.readAll(); err != nil {
			return err
		}
	}
	return json.Unmarshal(r.body, v)
}

// readAll drains all the HTTP response body read stream and then closes it.
func (r *Response) readAll() error {
	var err error
	r.body, err = io.ReadAll(r.resp.Body)
	if err != nil {
		return err
	}
	defer r.Close()
	return nil
}

// Close closes the HTTP response body read stream.
func (r *Response) Close() error {
	r.bodyClosed = true
	return r.resp.Body.Close()
}

// Method returns the HTTP request method.
func (r *Response) Method() string {
	return r.resp.Request.Method
}

// URL returns the HTTP request URL string.
func (r *Response) URL() string {
	return r.resp.Request.URL.String()
}

// Headers maps header keys to values. If the response had multiple headers
// with the same key, they may be concatenated, with comma delimiters.
func (r *Response) Headers() http.Header {
	return r.resp.Header
}

// Cookies parses and returns the cookies set in the Set-Cookie headers.
func (r *Response) Cookies() map[string]*http.Cookie {
	m := make(map[string]*http.Cookie)
	for _, c := range r.resp.Cookies() {
		m[c.Name] = c
	}
	return m
}
