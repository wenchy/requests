package requests

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// Response is a wrapper of http.Response.
type Response struct {
	*http.Response
	body []byte // auto filled from Response.Body
}

// newResponse reads and closes Response.Body. Then check the HTTP status
// in Response.StatusCode. It will return an error with status and text
// body embedded if status code is not 2xx, and none-nil response is also
// returned.
func newResponse(resp *http.Response, opts *Options) (*Response, error) {
	r := &Response{
		Response: resp,
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
func (r *Response) readAndCloseBody() (err error) {
	defer func() {
		err1 := r.Response.Body.Close()
		err = errors.Join(err, err1)
	}()
	r.body, err = io.ReadAll(r.Response.Body)
	return err
}

// StatusCode returns status code of HTTP response.
//
// NOTE: It returns -1 if response is nil.
func (r *Response) StatusCode() int {
	if r == nil || r.Response == nil {
		// return special status code -1 which is not registered with IANA.
		return -1
	}
	return r.Response.StatusCode
}

// StatusText returns a text for the HTTP status code.
//
// NOTE:
//   - It returns "Response is nil" if response is nil.
//   - It returns the empty string if the code is unknown.
//
// e.g. "OK"
func (r *Response) StatusText() string {
	if r == nil || r.Response == nil {
		// return special status code -1 which is not registered with IANA.
		return "Response is nil"
	}
	return r.Response.Status
}

// Bytes returns the HTTP response body as []byte.
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
	return r.Response.Request.Method
}

// URL returns the HTTP request URL string.
func (r *Response) URL() string {
	return r.Response.Request.URL.String()
}

// Headers maps header keys to values. If the response had multiple headers
// with the same key, they may be concatenated, with comma delimiters.
func (r *Response) Headers() http.Header {
	return r.Response.Header
}

// Cookies parses and returns the cookies set in the Set-Cookie headers.
func (r *Response) Cookies() map[string]*http.Cookie {
	m := make(map[string]*http.Cookie)
	for _, c := range r.Response.Cookies() {
		m[c.Name] = c
	}
	return m
}
