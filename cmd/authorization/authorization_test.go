package authorization

import (
	"strings"
	"testing"

	"github.com/h1067675/shortUrl/internal/logger"
)

func TestSetToken(t *testing.T) {
	type args struct {
		id int
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Set token ",
			args: args{
				id: 1,
			},
			want:    ".",
			wantErr: false,
		},
	}
	logger.Initialize("debug")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SetToken(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !strings.Contains(got, tt.want) {
				t.Errorf("SetToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckToken(t *testing.T) {
	type args struct {
		id int
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "Set token ",
			args: args{
				id: 1,
			},
			want:    1,
			wantErr: false,
		},
	}
	logger.Initialize("debug")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, _ := SetToken(tt.args.id)
			got, err := CheckToken(token)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CheckToken() = %v, want %v", got, tt.want)
			}
		})
	}
}
