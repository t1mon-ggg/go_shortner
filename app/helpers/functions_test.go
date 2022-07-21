package helpers

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t1mon-ggg/go_shortner/app/models"
)

func TestFanOut(t *testing.T) {
	type args struct {
		workers int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "create workers",
			args: args{
				workers: 10,
			},
			want: 10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FanOut(make(<-chan models.DelWorker), tt.args.workers)
			require.Equal(t, tt.want, len(got))
		})
	}
}
