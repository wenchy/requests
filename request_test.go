package requests

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pkg/errors"
)

func TestGet(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method is not GET: %s", r.Method)
		}
		t.Logf("query strings: %v", r.URL.Query())
		t.Logf("headers: %v", r.Header)
		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()
	type args struct {
		url     string
		options []Option
		timeout time.Duration
	}
	tests := []struct {
		name    string
		args    args
		want    *Response
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "test case 1",
			args: args{
				url: "https://www.google.com",
				options: []Option{
					BasicAuth("XXX", "OOO"),
				},
				timeout: 5 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "test case 2",
			args: args{
				url: "https://127.0.0.1:4004",
				options: []Option{
					Timeout(120 * time.Second),
				},
			},
			wantErr: true,
		},
		{
			name: "test case 3",
			args: args{
				url: testServer.URL,
				options: []Option{
					ParamPairs("param1", "value1"),
					ParamPairs("param2", "value2"),
					HeaderPairs("header1", "value1"),
					HeaderPairs("header2", "value2"),
				},
				timeout: 5 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "disable keep alive",
			args: args{
				url: testServer.URL,
				options: []Option{
					ParamPairs("param1", "value1"),
					ParamPairs("param2", "value2"),
					HeaderPairs("header1", "value1"),
					HeaderPairs("header2", "value2"),
					DisableKeepAlives(),
				},
				timeout: 5 * time.Second,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.args.timeout != 0 {
				SetEnvTimeout(tt.args.timeout)
			}
			got, err := Get(tt.args.url, tt.args.options...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != nil {
				fmt.Printf("status code: %v\n", got.StatusCode())
				fmt.Printf("headers: %+v\n", got.Headers())
				fmt.Printf("cookies: %+v\n", got.Cookies())
			} else {
				fmt.Printf("Get failed: %v\n", err)
			}
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("Get() = %v, want %v", got, tt.want)
			// }
		})
	}
}

func TestPost(t *testing.T) {
	filename1 := "./testdata/file1.txt"
	filename2 := "./testdata/file2.txt"
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleUpload := func(formKey string) error {
			// Go 1.17: net/http: multipart form should not include directory path in filename
			// Refer: https://github.com/golang/go/issues/45789
			file, header, err := r.FormFile(formKey)
			if err != nil {
				return errors.Wrapf(err, "get form file: %s failed", formKey)
			}
			defer file.Close()
			got, err := io.ReadAll(file)
			if err != nil {
				return errors.Wrap(err, "read all failed")
			}
			path := filepath.Join("./testdata/", header.Filename)
			src, err := os.ReadFile(path)
			if err != nil {
				return errors.Wrapf(err, "read file: %s failed", path)
			}
			diff := bytes.Compare(got, src)
			if diff != 0 {
				errors.Errorf("inconsistent content, expect: %s, got: %s", string(src), string(got))
			}
			return nil
		}

		if err := handleUpload("file1"); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			t.Errorf("handle upload failed: %+v", err)
			return
		}

		if err := handleUpload("file2"); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("upload form: file2 failed"))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("upload file success"))
	}))
	defer testServer.Close()

	fh1, err := os.Open(filename1)
	if err != nil {
		t.Errorf("open file: %s failed: %+v", filename1, err)
		return
	}
	defer fh1.Close()

	fh2, err := os.Open(filename2)
	if err != nil {
		t.Errorf("open file: %s failed: %+v", filename2, err)
		return
	}
	defer fh2.Close()

	type args struct {
		urlStr  string
		options []Option
	}
	tests := []struct {
		name    string
		args    args
		want    *Response
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "upload file test case 1",
			args: args{
				urlStr: testServer.URL,
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
				urlStr: "http://127.0.0.1:11111/unknown",
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
			resp, err := Post(tt.args.urlStr, tt.args.options...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Post() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if resp != nil {
				text, err := resp.Text()
				if (err != nil) != tt.wantErr {
					t.Errorf("Patch() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				t.Logf("resp: %s", text)
			}
		})
	}
}

func TestPatch(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("method is not PATCH: %s", r.Method)
		}
		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read body failed: %+v", err)
		}
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(b); err != nil {
			t.Errorf("write response failed: %+v", err)
		}
	}))
	defer testServer.Close()
	type args struct {
		urlStr  string
		options []Option
	}
	tests := []struct {
		name    string
		args    args
		want    *Response
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "patch test case 1",
			args: args{
				urlStr: testServer.URL,
				options: []Option{
					JSON(map[string]interface{}{
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
				urlStr: "http://127.0.0.1:11111/unknown",
				options: []Option{
					JSON(map[string]interface{}{
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
			resp, err := Patch(tt.args.urlStr, tt.args.options...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Patch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if resp != nil {
				text, err := resp.Text()
				if (err != nil) != tt.wantErr {
					t.Errorf("Patch() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				t.Logf("resp: %s", text)
			}
		})
	}
}
