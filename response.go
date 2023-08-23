package requests

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// Response is a wrapper of HTTP response.
type Response struct {
	resp *http.Response
	body []byte // auto filled from resp.Body
}

// newResponse reads and closes resp.Body. Then check the HTTP status
// in Response.StatusCode. It will return an error with status and text
// body embedded if status code is not 2xx, and none-nil response is also
// returned.
func newResponse(resp *http.Response, opts *httpOptions) (*Response, error) {
	r := &Response{
		resp: resp,
	}
	if err := r.readAndCloseBody(); err != nil {
		return nil, err
	}
	// return error with status and text body embedded if status code
	// is not 2xx, and response is also returned.
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		// TODO: only extracts 128 bytes from body.
		return r, errors.New(resp.Status + " " + r.Text())
	}
	if opts.ToText != nil {
		*opts.ToText = r.Text()
	}
	if opts.ToJSON != nil {
		if err := r.JSON(opts.ToJSON); err != nil {
			return r, err
		}
	}
	return r, nil
}

// readAndCloseBody drains all the HTTP response body stream and then closes it.
func (r *Response) readAndCloseBody() error {
	defer r.resp.Body.Close()
	var err error
	r.body, err = io.ReadAll(r.resp.Body)
	if err != nil {
		return err
	}
	return nil
}

// StatusCode returns status code of HTTP response.
func (r *Response) StatusCode() int {
	if r.resp == nil {
		return http.StatusServiceUnavailable
	}
	return r.resp.StatusCode
}

// Bytes parses the HTTP response body as []byte.
func (r *Response) Bytes() []byte {
	return r.body
}

// Text parses the HTTP response body as string.
func (r *Response) Text() string {
	return string(r.body)
}

// JSON decodes the HTTP response body as JSON format.
func (r *Response) JSON(v any) error {
	return json.Unmarshal(r.body, v)
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
