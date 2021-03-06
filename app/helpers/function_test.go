package helpers

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/t1mon-ggg/go_shortner/app/models"
)

func Test_mergeURLs(t *testing.T) {
	old := []models.ShortData{
		{Short: "One", Long: "One"},
		{Short: "Two", Long: "Two"},
	}
	new := []models.ShortData{
		{Short: "Two", Long: "Two"},
		{Short: "Three", Long: "Three"},
	}
	result := []models.ShortData{
		{Short: "One", Long: "One"},
		{Short: "Two", Long: "Two"},
		{Short: "Three", Long: "Three"},
	}
	data := mergeURLs(old, new)
	require.Equal(t, result, data)
}

func Test_mergeData(t *testing.T) {
	old := []models.ClientData{
		{
			Cookie: "cookie1",
			Key:    "Key1",
			Short: []models.ShortData{
				{
					Short: "Short1",
					Long:  "Long1",
				},
			},
		},
		{
			Cookie: "cookie2",
			Key:    "Key2",
			Short: []models.ShortData{
				{
					Short: "Short2",
					Long:  "Long2",
				},
				{
					Short: "Short3",
					Long:  "Long3",
				},
			},
		},
		{
			Cookie: "cookie3",
			Key:    "Key3",
			Short: []models.ShortData{
				{
					Short: "Short4",
					Long:  "Long4",
				},
			},
		},
	}
	new := models.ClientData{
		Cookie: "cookie2",
		Key:    "Key2",
		Short: []models.ShortData{
			{
				Short: "Short5",
				Long:  "Long5",
			},
		},
	}
	result := []models.ClientData{
		{
			Cookie: "cookie1",
			Key:    "Key1",
			Short: []models.ShortData{
				{
					Short: "Short1",
					Long:  "Long1",
				},
			},
		},
		{
			Cookie: "cookie2",
			Key:    "Key2",
			Short: []models.ShortData{
				{
					Short: "Short2",
					Long:  "Long2",
				},
				{
					Short: "Short3",
					Long:  "Long3",
				},
				{
					Short: "Short5",
					Long:  "Long5",
				},
			},
		},
		{
			Cookie: "cookie3",
			Key:    "Key3",
			Short: []models.ShortData{
				{
					Short: "Short4",
					Long:  "Long4",
				},
			},
		},
	}
	data := mergeData(old, new)
	require.Equal(t, result, data)
}

func Test_Merger(t *testing.T) {
	old := []models.ClientData{
		{
			Cookie: "cookie1",
			Key:    "Key1",
			Short: []models.ShortData{
				{
					Short: "Short1",
					Long:  "Long1",
				},
			},
		},
		{
			Cookie: "cookie2",
			Key:    "Key2",
			Short: []models.ShortData{
				{
					Short: "Short2",
					Long:  "Long2",
				},
				{
					Short: "Short3",
					Long:  "Long3",
				},
			},
		},
		{
			Cookie: "cookie3",
			Key:    "Key3",
			Short: []models.ShortData{
				{
					Short: "Short4",
					Long:  "Long4",
				},
			},
		},
	}
	new1 := models.ClientData{
		Cookie: "cookie2",
		Key:    "Key2",
		Short: []models.ShortData{
			{
				Short: "Short5",
				Long:  "Long5",
			},
		},
	}
	new2 := models.ClientData{
		Cookie: "cookie3",
		Key:    "Key3",
		Short: []models.ShortData{
			{
				Short: "Short4",
				Long:  "Long4",
			},
		},
	}
	result := []models.ClientData{
		{
			Cookie: "cookie1",
			Key:    "Key1",
			Short: []models.ShortData{
				{
					Short: "Short1",
					Long:  "Long1",
				},
			},
		},
		{
			Cookie: "cookie2",
			Key:    "Key2",
			Short: []models.ShortData{
				{
					Short: "Short2",
					Long:  "Long2",
				},
				{
					Short: "Short3",
					Long:  "Long3",
				},
				{
					Short: "Short5",
					Long:  "Long5",
				},
			},
		},
		{
			Cookie: "cookie3",
			Key:    "Key3",
			Short: []models.ShortData{
				{
					Short: "Short4",
					Long:  "Long4",
				},
			},
		},
	}
	data, err := Merger(old, new1)
	require.Equal(t, result, data)
	require.NoError(t, err)
	_, err = Merger(old, new2)
	require.Error(t, err)
}

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
