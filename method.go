package requests

import "net/http"

// Get sends an HTTP request with GET method.
//
// On error, any Response can be ignored. A non-nil Response with a
// non-nil error only occurs when Response.StatusCode() is not 2xx.
func Get(url string, options ...Option) (*Response, error) {
	return callMethod(http.MethodGet, url, options...)
}

// Post sends an HTTP POST request.
func Post(url string, options ...Option) (*Response, error) {
	return callMethod(http.MethodPost, url, options...)
}

// Put sends an HTTP request with PUT method.
//
// On error, any Response can be ignored. A non-nil Response with a
// non-nil error only occurs when Response.StatusCode() is not 2xx.
func Put(url string, options ...Option) (*Response, error) {
	return callMethod(http.MethodPut, url, options...)
}

// Patch sends an HTTP request with PATCH method.
//
// On error, any Response can be ignored. A non-nil Response with a
// non-nil error only occurs when Response.StatusCode() is not 2xx.
func Patch(url string, options ...Option) (*Response, error) {
	return callMethod(http.MethodPatch, url, options...)
}

// Delete sends an HTTP request with DELETE method.
//
// On error, any Response can be ignored. A non-nil Response with a
// non-nil error only occurs when Response.StatusCode() is not 2xx.
func Delete(url string, options ...Option) (*Response, error) {
	return callMethod(http.MethodDelete, url, options...)
}
