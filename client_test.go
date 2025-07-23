package requests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClientOption(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("query strings: %v", r.URL.Query())
		t.Logf("headers: %v", r.Header)
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()
	type args struct {
		url     string
		options []ClientOption
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "with timeout",
			args: args{
				url: testServer.URL,
				options: []ClientOption{
					WithTimeout(time.Millisecond),
				},
			},
			wantErr: true,
		},
		{
			name: "with transport",
			args: args{
				url: testServer.URL,
				options: []ClientOption{
					WithTransport(func() http.RoundTripper {
						transport := http.DefaultTransport.(*http.Transport).Clone()
						transport.DisableKeepAlives = true
						return transport
					}()),
				},
			},
			wantErr: false,
		},
		{
			name: "with interceptor",
			args: args{
				url: testServer.URL,
				options: []ClientOption{
					WithInterceptor(func(ctx context.Context, r *Request, do Do) (*Response, error) {
						t.Logf("method: %s", r.Method)
						return do(ctx, r)
					}),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli := NewClient(tt.args.options...)
			got, err := cli.Get(tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				t.Logf("response body: %+v\n", got.Text())
			}
		})
	}
}
