package requests

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/http/httptrace"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func init() {
	InitDefaultClient(WithInterceptor(ChainInterceptors(logInterceptor, metricInterceptor, traceInterceptor)))
}

func logInterceptor(ctx context.Context, r *Request, do Do) (*Response, error) {
	log.Printf("method: %s", r.Method)
	return do(ctx, r)
}

func metricInterceptor(ctx context.Context, r *Request, do Do) (*Response, error) {
	log.Printf("request, method: %s, url: %s, bodySize: %d", r.Method, r.URL, len(r.Bytes()))
	resp, err := do(ctx, r)
	if err == nil {
		log.Printf("response: method: %s, status: %s, bodySize: %d", r.Method, resp.StatusText(), len(resp.Bytes()))
	}
	return resp, err
}

func traceInterceptor(ctx context.Context, r *Request, do Do) (*Response, error) {
	trace := &httptrace.ClientTrace{
		GetConn:      func(hostPort string) { log.Printf("starting to create conn: %s ", hostPort) },
		DNSStart:     func(info httptrace.DNSStartInfo) { log.Printf("starting to look up dns: %+v", info) },
		DNSDone:      func(info httptrace.DNSDoneInfo) { log.Printf("done looking up dns: %+v", info) },
		ConnectStart: func(network, addr string) { log.Printf("starting tcp connection: %s, %s", network, addr) },
		ConnectDone: func(network, addr string, err error) {
			log.Printf("tcp connection created: %s, %s, %v", network, addr, err)
		},
		GotConn: func(info httptrace.GotConnInfo) { log.Printf("connection established: %+v", info) },
	}
	ctx = httptrace.WithClientTrace(ctx, trace)
	return do(ctx, r)
}

type testRequest struct {
	Headers http.Header
	Params  url.Values
	Form    url.Values
}

func TestGet(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("query strings: %v", r.URL.Query())
		t.Logf("headers: %v", r.Header)
		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()
	type args struct {
		url     string
		options []Option
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "basic auth",
			args: args{
				url: testServer.URL,
				options: []Option{
					BasicAuth("XXX", "OOO"),
				},
			},
			wantErr: false,
		},
		{
			name: "http server not available",
			args: args{
				url: "https://127.0.0.1:4004",
				options: []Option{
					Timeout(120 * time.Second),
				},
			},
			wantErr: true,
		},
		{
			name: "manipulate URLs and query parameters",
			args: args{
				url: testServer.URL + "/get?a=1&b=2",
				options: []Option{
					ParamPairs("b", "20", "c", "30"),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Get(tt.args.url, tt.args.options...)
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

func TestGetWithContext(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()
	ctx10ms, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	ctx200ms, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	type args struct {
		url     string
		options []Option
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "with context 10ms",
			args: args{
				url: testServer.URL,
				options: []Option{
					Context(ctx10ms),
				},
			},
			wantErr: true,
		},
		{
			name: "with context 200ms",
			args: args{
				url: testServer.URL,
				options: []Option{
					Context(ctx200ms),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Get(tt.args.url, tt.args.options...)
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

func TestPostBody(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, r.Body.Close())
		}()
		w.WriteHeader(http.StatusOK)
		n, err := w.Write(body)
		assert.NoError(t, err)
		assert.Equal(t, n, len(body))
	}))
	defer testServer.Close()

	type args struct {
		url     string
		options []Option
	}
	tests := []struct {
		name    string
		args    args
		want    *Response
		wantErr bool
	}{
		{
			name: "io.Reader body",
			args: args{
				url: testServer.URL,
				options: []Option{
					Body(strings.NewReader("test1")),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Post(tt.args.url, tt.args.options...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != nil {
				t.Logf("status code: %+v", got.StatusCode())
				t.Logf("headers: %+v", got.Headers())
				t.Logf("cookies: %+v", got.Cookies())
				t.Logf("body: %+v", got.Text())
			} else {
				t.Logf("Get failed: %v", err)
			}
		})
	}
}

func TestPostData(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, r.Body.Close())
		}()
		w.WriteHeader(http.StatusOK)
		n, err := w.Write(body)
		assert.NoError(t, err)
		assert.Equal(t, n, len(body))
	}))
	defer testServer.Close()

	type args struct {
		url     string
		options []Option
	}
	tests := []struct {
		name    string
		args    args
		want    *Response
		wantErr bool
	}{
		{
			name: "data",
			args: args{
				url: testServer.URL,
				options: []Option{
					Data("test1"),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Post(tt.args.url, tt.args.options...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != nil {
				t.Logf("status code: %+v", got.StatusCode())
				t.Logf("headers: %+v", got.Headers())
				t.Logf("cookies: %+v", got.Cookies())
				t.Logf("body: %+v", got.Text())
			} else {
				t.Logf("Get failed: %v", err)
			}
		})
	}
}

func TestPostForm(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		assert.NoError(t, err)
		req := testRequest{
			Headers: r.Header,
			Params:  r.URL.Query(),
			Form:    r.Form,
		}
		t.Logf("query strings: %v", r.URL.Query())
		t.Logf("headers: %v", r.Header)
		w.WriteHeader(http.StatusOK)
		data, err := json.Marshal(req)
		assert.NoError(t, err)
		n, err := w.Write(data)
		assert.NoError(t, err)
		assert.Equal(t, n, len(data))
	}))
	defer testServer.Close()
	type args struct {
		url     string
		options []Option
	}
	tests := []struct {
		name    string
		args    args
		want    *testRequest
		wantErr bool
	}{
		{
			name: "headers",
			args: args{
				url: testServer.URL,
				options: []Option{
					Headers(map[string]string{"header1": "value1"}),
					Headers(http.Header{"header2": []string{"value2", "value2-2"}}),
					HeaderPairs("header1", "value1-2"),
					HeaderPairs("header2", "value2-3", "header2", "value2-4"),
				},
			},
			want: &testRequest{
				Headers: http.Header{
					http.CanonicalHeaderKey("header1"): []string{"value1", "value1-2"},
					http.CanonicalHeaderKey("header2"): []string{"value2", "value2-2", "value2-3", "value2-4"},
				},
			},
			wantErr: false,
		},
		{
			name: "params",
			args: args{
				url: testServer.URL,
				options: []Option{
					Params(map[string]string{"param1": "value1"}),
					Params(url.Values{"param2": []string{"value2", "value2-2"}}),
					ParamPairs("param1", "value1-2"),
					ParamPairs("param2", "value2-3", "param2", "value2-4"),
				},
			},
			want: &testRequest{
				Params: url.Values{
					"param1": []string{"value1", "value1-2"},
					"param2": []string{"value2", "value2-2", "value2-3", "value2-4"},
				},
			},
			wantErr: false,
		},
		{
			name: "form",
			args: args{
				url: testServer.URL,
				options: []Option{
					Form(map[string]string{"form1": "value1"}),
					Form(url.Values{"form2": []string{"value2", "value2-2"}}),
					FormPairs("form1", "value1-2"),
					FormPairs("form2", "value2-3", "form2", "value2-4"),
				},
			},
			want: &testRequest{
				Form: url.Values{
					"form1": []string{"value1", "value1-2"},
					"form2": []string{"value2", "value2-2", "value2-3", "value2-4"},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Post(tt.args.url, tt.args.options...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && tt.want != nil {
				rsp := &testRequest{}
				err := got.JSON(rsp)
				assert.NoError(t, err)
				assert.Subsetf(t, rsp.Headers, tt.want.Headers, "some headers missing in HTTP server-side")
				assert.Subsetf(t, rsp.Params, tt.want.Params, "some params missing in HTTP server-side")
				assert.Subsetf(t, rsp.Form, tt.want.Form, "some form data missing in HTTP server-side")
				t.Logf("got testRequest: %+v\n", rsp)
			}
		})
	}
}

type EchoRequest struct {
	ID   uint32
	Name string
}

type EchoResponse struct {
	ID   uint32
	Name string
}

func TestPostJSON(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method is not Post: %s", r.Method)
		}
		t.Logf("query strings: %v", r.URL.Query())
		t.Logf("headers: %v", r.Header)

		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, r.Body.Close())
		}()
		var req EchoRequest
		err = json.Unmarshal(body, &req)
		assert.NoError(t, err)

		jsonResp := &EchoResponse{
			ID:   req.ID,
			Name: "echo " + req.Name,
		}
		respBytes, err := json.Marshal(jsonResp)
		assert.NoError(t, err)
		w.WriteHeader(http.StatusOK)
		n, err := w.Write(respBytes)
		assert.NoError(t, err)
		assert.Equal(t, n, len(respBytes))
	}))
	defer testServer.Close()

	var jsonResp EchoResponse
	var textResp string
	var reqDump, respDump string
	type args struct {
		url     string
		options []Option
	}
	tests := []struct {
		name    string
		args    args
		want    *Response
		wantErr bool
	}{
		{
			name: "json-request-and-response",
			args: args{
				url: testServer.URL,
				options: []Option{
					ParamPairs("param1", "value1"),
					ParamPairs("param2", "value2"),
					HeaderPairs("header1", "value1"),
					HeaderPairs("header2", "value2"),
					JSON(&EchoRequest{ID: 1, Name: "Hello"}),
					ToJSON(&jsonResp),
					ToText(&textResp),
					Dump(&reqDump, &respDump),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Post(tt.args.url, tt.args.options...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != nil {
				t.Logf("status code: %+v", got.StatusCode())
				t.Logf("headers: %+v", got.Headers())
				t.Logf("cookies: %+v", got.Cookies())
				t.Logf("body: %+v", got.Text())
				t.Logf("body(text): %+v", textResp)
				t.Logf("body(json): %+v", jsonResp)
				t.Logf("Request(dump):\n%s", reqDump)
				t.Logf("Response(dump):\n%s", respDump)
			} else {
				t.Logf("Get failed: %v", err)
			}
		})
	}
}

func TestPostFiles(t *testing.T) {
	filename1 := "./testdata/file1.txt"
	filename2 := "./testdata/file2.txt"
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleUpload := func(formKey string) error {
			// Go 1.17: net/http: multipart form should not include directory path in filename
			// Refer: https://github.com/golang/go/issues/45789
			file, header, err := r.FormFile(formKey)
			assert.NoErrorf(t, err, "get form file: %s failed", formKey)
			defer func() {
				assert.NoError(t, file.Close())
			}()
			got, err := io.ReadAll(file)
			assert.NoError(t, err)
			path := filepath.Join("./testdata/", header.Filename)
			src, err := os.ReadFile(path)
			assert.NoError(t, err)

			assert.Equalf(t, string(src), string(got), "content not same: %s", formKey)
			return nil
		}

		if err := handleUpload("file1"); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			t.Errorf("handle upload failed: %+v", err)
			return
		}

		if err := handleUpload("file2"); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			failedMsg := "upload form: file2 failed"
			n, err := w.Write([]byte(failedMsg))
			assert.NoError(t, err)
			assert.Equal(t, n, len(failedMsg))
			return
		}

		w.WriteHeader(http.StatusOK)
		successMsg := "upload file success"
		n, err := w.Write([]byte(successMsg))
		assert.NoError(t, err)
		assert.Equal(t, n, len(successMsg))
	}))
	defer testServer.Close()

	fh1, err := os.Open(filename1)
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, fh1.Close())
	}()

	fh2, err := os.Open(filename2)
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, fh2.Close())
	}()

	type args struct {
		url     string
		options []Option
	}
	tests := []struct {
		name    string
		args    args
		want    *Response
		wantErr bool
	}{
		{
			name: "upload file test case 1",
			args: args{
				url: testServer.URL,
				options: []Option{
					Files(map[string]*os.File{
						"file1": fh1,
						"file2": fh2,
					}),
					Timeout(120 * time.Second),
				},
			},
			wantErr: false,
		},
		{
			name: "upload file test case 2",
			args: args{
				url: "http://127.0.0.1:11111/unknown",
				options: []Option{
					Files(map[string]*os.File{
						"file1": fh1,
						"file2": fh2,
					}),
					Timeout(120 * time.Second),
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var dump string
			resp, err := Post(tt.args.url, append(tt.args.options, Dump(&dump, nil))...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Post() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if resp != nil {
				t.Logf("resp: %s", resp.Text())
			}
		})
	}
}

func TestPatch(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equalf(t, http.MethodPatch, r.Method, "method is not PATCH: %s", r.Method)
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		w.WriteHeader(http.StatusOK)
		n, err := w.Write(body)
		assert.NoError(t, err)
		assert.Equal(t, n, len(body))
	}))
	defer testServer.Close()
	type args struct {
		url     string
		options []Option
	}
	tests := []struct {
		name    string
		args    args
		want    *Response
		wantErr bool
	}{
		{
			name: "patch test case 1",
			args: args{
				url: testServer.URL,
				options: []Option{
					JSON(map[string]any{
						"status":  0,
						"message": "hello http patch",
					}),
					Timeout(120 * time.Second),
				},
			},
			wantErr: false,
		},
		{
			name: "patch test case 2",
			args: args{
				url: "http://127.0.0.1:11111/unknown",
				options: []Option{
					JSON(map[string]any{
						"status":  0,
						"message": "hello http patch",
					}),
					Timeout(120 * time.Second),
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var dump string
			resp, err := Patch(tt.args.url, append(tt.args.options, Dump(&dump, nil))...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Patch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if resp != nil {
				t.Logf("resp: %s", resp.Text())
			}
		})
	}
}

func TestInterceptors(t *testing.T) {
	filename1 := "./testdata/file1.txt"
	filename2 := "./testdata/file2.txt"
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, r.Body.Close())
		}()

		assert.Equalf(t, strconv.Itoa(len(body)), r.Header.Get("X-Body-Size"), "content length not same")
		assert.Equalf(t, hex.EncodeToString(md5.New().Sum(body)), r.Header.Get("X-Body-Md5"), "content md5 not same")
	}))
	defer testServer.Close()

	fh1, err := os.Open(filename1)
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, fh1.Close())
	}()

	fh2, err := os.Open(filename2)
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, fh2.Close())
	}()

	type args struct {
		url     string
		options []Option
	}
	tests := []struct {
		name    string
		args    args
		want    *Response
		wantErr bool
	}{
		{
			name: "body",
			args: args{
				url: testServer.URL,
				options: []Option{
					Body(strings.NewReader("test1")),
				},
			},
			wantErr: false,
		},
		{
			name: "data",
			args: args{
				url: testServer.URL,
				options: []Option{
					Data("test1"),
				},
			},
			wantErr: false,
		},
		{
			name: "form",
			args: args{
				url: testServer.URL,
				options: []Option{
					Form(map[string]string{"form1": "value1"}),
					Form(url.Values{"form2": []string{"value2", "value2-2"}}),
					FormPairs("form1", "value1-2"),
					FormPairs("form2", "value2-3", "form2", "value2-4"),
				},
			},
			wantErr: false,
		},
		{
			name: "json",
			args: args{
				url: testServer.URL,
				options: []Option{
					ParamPairs("param1", "value1"),
					ParamPairs("param2", "value2"),
					HeaderPairs("header1", "value1"),
					HeaderPairs("header2", "value2"),
					JSON(&EchoRequest{ID: 1, Name: "Hello"}),
				},
			},
			wantErr: false,
		},
		{
			name: "file",
			args: args{
				url: testServer.URL,
				options: []Option{
					Files(map[string]*os.File{
						"file1": fh1,
						"file2": fh2,
					}),
					Timeout(120 * time.Second),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := Post(tt.args.url, append(tt.args.options, Interceptor(func(ctx context.Context, req *Request, do Do) (*Response, error) {
				req.Header.Set("X-Body-Size", strconv.Itoa(len(req.Bytes())))
				req.Header.Set("X-Body-Md5", hex.EncodeToString(md5.New().Sum(req.Bytes())))
				return do(ctx, req)
			}))...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Post() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if resp != nil {
				t.Logf("resp: %s", resp.Text())
			}
		})
	}
}
