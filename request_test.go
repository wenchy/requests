package requests

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/pkg/errors"
)

func TestGet(t *testing.T) {
	type args struct {
		url     string
		setters []Option
		timeout int64
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
				setters: []Option{
					BasicAuth("XXX", "OOO"),
				},
				timeout: 5,
			},
			wantErr: false,
		},
		{
			name: "test case 2",
			args: args{
				url: "https://127.0.0.1:4004",
				setters: []Option{
					Timeout(120),
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.args.timeout != 0 {
				SetEnvTimeout(tt.args.timeout)
			}
			got, err := Get(tt.args.url, tt.args.setters...)
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
	filename1 := "./test/data/file1.txt"
	filename2 := "./test/data/file2.txt"
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleUpload := func(formKey string) error {
			file, header, err := r.FormFile(formKey)
			if err != nil {
				return errors.Wrapf(err, "get form file: %s failed", formKey)
			}
			defer file.Close()
			got, err := ioutil.ReadAll(file)
			if err != nil {
				return errors.Wrap(err, "read all failed")
			}
			src, err := ioutil.ReadFile(header.Filename)
			if err != nil {
				return errors.Wrapf(err, "read file: %s failed", header.Filename)
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
	defer fh1.Close()
	if err != nil {
		t.Errorf("open file: %s failed: %+v", filename1, err)
		return
	}

	fh2, err := os.Open(filename2)
	defer fh2.Close()
	if err != nil {
		t.Errorf("open file: %s failed: %+v", filename2, err)
		return
	}

	type args struct {
		rawurl  string
		setters []Option
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
				rawurl: testServer.URL,
				setters: []Option{
					Files(map[string]*os.File{
						"file1": fh1,
						"file2": fh2,
					}),
					Timeout(120),
				},
			},
			wantErr: false,
		},
		{
			name: "upload file test case 2",
			args: args{
				rawurl: "http://127.0.0.1:11111/unknown",
				setters: []Option{
					Files(map[string]*os.File{
						"file1": fh1,
						"file2": fh2,
					}),
					Timeout(120),
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := Post(tt.args.rawurl, tt.args.setters...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Post() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Logf("%+v\n", resp)
		})
	}
}
