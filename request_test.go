package requests

import (
	"fmt"
	"testing"
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
			} else {
				fmt.Printf("Get failed: %v\n", err)
			}
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("Get() = %v, want %v", got, tt.want)
			// }
		})
	}
}
