package requests

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
)

// Response is a wrapper of http response.
type Response struct {
	rsp        *http.Response
	body       []byte // filled if rsp.Body was already drained.
	bodyClosed bool
}

// StatusCode get status code of http response.
func (r *Response) StatusCode() int {
	if r.rsp == nil {
		return http.StatusServiceUnavailable
	}
	return r.rsp.StatusCode
}

// Raw get the raw socket response from the server.
func (r *Response) Raw() io.ReadCloser {
	return r.rsp.Body
}

// Text return the http response body as string.
func (r *Response) Text() (string, error) {
	if !r.bodyClosed {
		if err := r.readAll(); err != nil {
			return "", err
		}
	}

	return string(r.body), nil
}

// JSON decode the http response body as JSON format.
func (r *Response) JSON(v interface{}) error {
	if !r.bodyClosed {
		if err := r.readAll(); err != nil {
			return err
		}
	}
	return json.Unmarshal(r.body, v)
}

// readAll drain all the http response body read stream and close the stream.
func (r *Response) readAll() error {
	var err error
	r.body, err = ioutil.ReadAll(r.rsp.Body)
	if err != nil {
		return err
	}
	defer r.Close()

	return nil
}

// Close closes the http response body read stream.
func (r *Response) Close() error {
	r.bodyClosed = true
	return r.rsp.Body.Close()
}

// Method get the http request method.
func (r *Response) Method() string {
	return r.rsp.Request.Method
}

// URL get the http request url string.
func (r *Response) URL() string {
	return r.rsp.Request.URL.String()
}

func (r *Response) Header(key string) string {
	return r.rsp.Header.Get(key)
}

func (r *Response) Cookies() []*http.Cookie {
	return r.rsp.Cookies()
}
