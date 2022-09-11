package putio

import (
	"fmt"
	"testing"

	"github.com/putdotio/go-putio"
)

func TestIsNotFound(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "is not found error",
			args: args{&putio.ErrorResponse{Type: "NotFound"}},
			want: true,
		},
		{
			name: "is not not found error",
			args: args{fmt.Errorf("NotFound")},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNotFound(tt.args.err); got != tt.want {
				t.Errorf("IsNotFound() = %v, want %v", got, tt.want)
			}
		})
	}
}
