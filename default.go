package requests

import (
	"net/http"
	"sync"

	"github.com/Wenchy/requests/internal/auth/redirector"
)

var (
	once          sync.Once
	defaultClient *Client
)

func newDefaultClient() *Client {
	return &Client{
		Client: &http.Client{
			CheckRedirect: redirector.RedirectPolicyFunc,
		},
	}
}

func GetDefaultClient() *Client {
	once.Do(func() {
		defaultClient = newDefaultClient()
	})
	return defaultClient
}

func InitDefaultClient(options ...ClientOption) {
	client := GetDefaultClient()
	for _, setter := range options {
		setter(client)
	}
}
