package requests

import (
	"net/http"
	"sync"
	"time"

	"github.com/Wenchy/requests/internal/auth/redirector"
)

var (
	once          sync.Once
	defaultClient *Client
)

func newDefaultClient() *Client {
	return &Client{
		client: &http.Client{
			CheckRedirect: redirector.RedirectPolicyFunc,
			Timeout:       10 * time.Second,
		},
	}
}

func getDefaultClient() *Client {
	once.Do(func() {
		defaultClient = newDefaultClient()
	})
	return defaultClient
}

// InitDefaultClient initializes the default client with given options.
func InitDefaultClient(setters ...ClientOption) {
	client := getDefaultClient()
	for _, setter := range setters {
		setter(client)
	}
}
