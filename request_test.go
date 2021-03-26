package requests

import (
	"fmt"
	"testing"
)

func TestGet(t *testing.T) {
	type args struct {
		url     string
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
			name: "test case 1",
			args: args{
				url:     "http://9.134.55.198/dev/",
				setters: []Option{
					BasicAuth("wenchyzhu", "wenchyzhu"),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Get(tt.args.url, tt.args.setters...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			fmt.Printf("status code: %v\n", got.StatusCode())
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("Get() = %v, want %v", got, tt.want)
			// }
		})
	}
}
