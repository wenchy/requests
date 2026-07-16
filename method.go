package requests

// Get sends an HTTP GET request.
//
// On error, any Response can be ignored. A non-nil Response with a
// non-nil error only occurs when Response.StatusCode() is not 2xx.
func Get(url string, options ...Option) (*Response, error) {
	return getDefaultClient().Get(url, options...)
}

// Post sends an HTTP POST request.
func Post(url string, options ...Option) (*Response, error) {
	return getDefaultClient().Post(url, options...)
}

// Put sends an HTTP PUT request.
//
// On error, any Response can be ignored. A non-nil Response with a
// non-nil error only occurs when Response.StatusCode() is not 2xx.
func Put(url string, options ...Option) (*Response, error) {
	return getDefaultClient().Put(url, options...)
}

// Patch sends an HTTP PATCH request.
//
// On error, any Response can be ignored. A non-nil Response with a
// non-nil error only occurs when Response.StatusCode() is not 2xx.
func Patch(url string, options ...Option) (*Response, error) {
	return getDefaultClient().Patch(url, options...)
}

// Delete sends an HTTP DELETE request.
//
// On error, any Response can be ignored. A non-nil Response with a
// non-nil error only occurs when Response.StatusCode() is not 2xx.
func Delete(url string, options ...Option) (*Response, error) {
	return getDefaultClient().Delete(url, options...)
}
