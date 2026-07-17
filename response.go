package requests

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// Response wraps [http.Response].
type Response struct {
	*http.Response
	body []byte // auto filled from Response.Body
}

// newResponse reads and closes the response body. It returns a non-nil
// response along with an error containing the status and text body when
// the status code is not 2xx.
func newResponse(resp *http.Response, opts *Options) (*Response, error) {
	r := &Response{
		Response: resp,
	}
	if err := r.readAndCloseBody(); err != nil {
		return nil, err
	}
	// non-2xx: return the response along with an error.
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

// readAndCloseBody drains and closes the response body.
func (r *Response) readAndCloseBody() (err error) {
	defer func() {
		err1 := r.Response.Body.Close()
		err = errors.Join(err, err1)
	}()
	r.body, err = io.ReadAll(r.Response.Body)
	return err
}

// StatusCode returns the HTTP response status code, or -1 if the response is nil.
func (r *Response) StatusCode() int {
	if r == nil || r.Response == nil {
		// return special status code -1 which is not registered with IANA.
		return -1
	}
	return r.Response.StatusCode
}

// StatusText returns the HTTP status text, or "<nil>" if the response is nil.
func (r *Response) StatusText() string {
	if r == nil || r.Response == nil {
		return "<nil>"
	}
	return r.Response.Status
}

// Bytes returns the HTTP response body as []byte.
func (r *Response) Bytes() []byte {
	return r.body
}

// Text returns the HTTP response body as a string.
func (r *Response) Text() string {
	return string(r.body)
}

// JSON decodes the HTTP response body into v as JSON.
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
	cookies := make(map[string]*http.Cookie)
	for _, c := range r.Response.Cookies() {
		cookies[c.Name] = c
	}
	return cookies
}
