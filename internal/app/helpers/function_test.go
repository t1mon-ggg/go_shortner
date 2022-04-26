package helpers

import (
	"testing"

	"github.com/stretchr/testify/require"

	. "github.com/t1mon-ggg/go_shortner/internal/app/models"
)

func Test_mergeURLs(t *testing.T) {
	old := map[string]string{
		"one": "one",
		"two": "two",
	}
	new := map[string]string{
		"two":   "two",
		"three": "three",
	}
	result := map[string]string{
		"one":   "one",
		"two":   "two",
		"three": "three",
	}
	data := mergeURLs(old, new)
	require.Equal(t, result, data)
}

func Test_mergeData(t *testing.T) {
	old := make(map[string]WebData)
	new := make(map[string]WebData)
	cookie1 := WebData{Key: "secret-key1",
		Short: map[string]string{
			"ABCDEFGH": "http://example1.org",
			"12345678": "http://example2.org",
		},
	}
	cookie2 := WebData{
		Key: "secret-key2",
		Short: map[string]string{
			"87654321": "http://example3.org",
		},
	}
	cookie4 := WebData{
		Key: "secret-key4",
		Short: map[string]string{
			"87654321": "http://example7.org",
		},
	}
	old["cookie1"] = cookie1
	old["cookie2"] = cookie2
	old["cookie4"] = cookie4
	newcookie1 := WebData{
		Key: "secret-key1",
		Short: map[string]string{
			"abcdABCD": "http://example4.org",
		},
	}
	newcookie2 := WebData{
		Key: "",
		Short: map[string]string{
			"AbCdAbCd": "http://example5.org",
		},
	}
	newcookie3 := WebData{
		Key: "secret-key3",
		Short: map[string]string{
			"AbCdAbCd": "http://example6.org",
		},
	}
	newcookie5 := WebData{
		Key:   "secret-key5",
		Short: map[string]string{},
	}
	new["cookie1"] = newcookie1
	new["cookie2"] = newcookie2
	new["cookie3"] = newcookie3
	new["cookie5"] = newcookie5
	result := make(map[string]WebData)
	rcookie1 := WebData{
		Key: "secret-key1",
		Short: map[string]string{
			"ABCDEFGH": "http://example1.org",
			"12345678": "http://example2.org",
			"abcdABCD": "http://example4.org",
		},
	}
	rcookie2 := WebData{
		Key: "secret-key2",
		Short: map[string]string{
			"87654321": "http://example3.org",
			"AbCdAbCd": "http://example5.org",
		},
	}
	rcookie3 := WebData{
		Key: "secret-key3",
		Short: map[string]string{
			"AbCdAbCd": "http://example6.org",
		},
	}
	rcookie4 := WebData{
		Key: "secret-key4",
		Short: map[string]string{
			"87654321": "http://example7.org",
		},
	}
	result["cookie1"] = rcookie1
	result["cookie2"] = rcookie2
	result["cookie3"] = rcookie3
	result["cookie4"] = rcookie4
	result["cookie5"] = newcookie5
	data := mergeData(old, new)
	require.Equal(t, result, data)
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
