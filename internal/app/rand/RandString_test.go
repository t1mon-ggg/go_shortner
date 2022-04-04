package rand

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRandStringRunes(t *testing.T) {
	tests := []struct {
		name string
		n    int
		want int
	}{
		{
			name: "Generate 8 symbols",
			n:    8,
			want: 8,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RandStringRunes(tt.n)
			require.Equal(t, tt.want, len(got))
		})
	}
}
