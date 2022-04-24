package storage

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/t1mon-ggg/go_shortner/internal/app/helpers"
)

func Test_MEM_Write(t *testing.T) {
	db := NewMemDB()
	data := make(map[string]helpers.WebData)
	data["cookie1"] = helpers.WebData{Key: "secret_key", Short: map[string]string{"abcdABCD": "http://example.org"}}
	err := db.Write(data)
	require.NoError(t, err)
}

func Test_MEM_ReadByCookie(t *testing.T) {
	db := NewMemDB()
	data := make(map[string]helpers.WebData)
	data["cookie1"] = helpers.WebData{Key: "secret_key1", Short: map[string]string{"abcdABC1": "http://example1.org"}}
	data["cookie2"] = helpers.WebData{Key: "secret_key2", Short: map[string]string{"abcdABC2": "http://example2.org"}}
	data["cookie3"] = helpers.WebData{Key: "secret_key3", Short: map[string]string{"abcdABC3": "http://example3.org"}}
	err := db.Write(data)
	require.NoError(t, err)
	expected := helpers.Data{"cookie2": helpers.WebData{Key: "secret_key2", Short: map[string]string{"abcdABC2": "http://example2.org"}}}
	val, err := db.ReadByCookie("cookie2")
	require.NoError(t, err)
	require.Equal(t, expected, val)
}

func Test_MEM_ReadByTag(t *testing.T) {
	db := NewMemDB()
	data := make(map[string]helpers.WebData)
	data["cookie1"] = helpers.WebData{Key: "secret_key1", Short: map[string]string{"abcdABC1": "http://example1.org"}}
	data["cookie2"] = helpers.WebData{Key: "secret_key2", Short: map[string]string{"abcdABC2": "http://example2.org"}}
	data["cookie3"] = helpers.WebData{Key: "secret_key3", Short: map[string]string{"abcdABC3": "http://example3.org"}}
	err := db.Write(data)
	require.NoError(t, err)
	expected := make(map[string]string)
	expected["abcdABC2"] = "http://example2.org"
	val, err := db.ReadByTag("abcdABC2")
	require.NoError(t, err)
	require.Equal(t, expected, val)
}

func Test_MEM_Close(t *testing.T) {
	db := NewMemDB()
	data := make(map[string]helpers.WebData)
	data["cookie1"] = helpers.WebData{Key: "secret_key1", Short: map[string]string{"abcdABC1": "http://example1.org"}}
	data["cookie2"] = helpers.WebData{Key: "secret_key2", Short: map[string]string{"abcdABC2": "http://example2.org"}}
	data["cookie3"] = helpers.WebData{Key: "secret_key3", Short: map[string]string{"abcdABC3": "http://example3.org"}}
	err := db.Write(data)
	require.NoError(t, err)
	err = db.Close()
	require.NoError(t, err)
	require.Equal(t, MemDB{}, db)
}

func Test_MEM_Ping(t *testing.T) {
	db := NewMemDB()
	data := make(map[string]helpers.WebData)
	data["cookie1"] = helpers.WebData{Key: "secret_key1", Short: map[string]string{"abcdABC1": "http://example1.org"}}
	data["cookie2"] = helpers.WebData{Key: "secret_key2", Short: map[string]string{"abcdABC2": "http://example2.org"}}
	data["cookie3"] = helpers.WebData{Key: "secret_key3", Short: map[string]string{"abcdABC3": "http://example3.org"}}
	err := db.Write(data)
	require.NoError(t, err)
	err = db.Ping()
	require.NoError(t, err)
}

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
	old := make(map[string]helpers.WebData)
	new := make(map[string]helpers.WebData)
	cookie1 := helpers.WebData{Key: "secret-key1",
		Short: map[string]string{
			"ABCDEFGH": "http://example1.org",
			"12345678": "http://example2.org",
		},
	}
	cookie2 := helpers.WebData{
		Key: "secret-key2",
		Short: map[string]string{
			"87654321": "http://example3.org",
		},
	}
	cookie4 := helpers.WebData{
		Key: "secret-key4",
		Short: map[string]string{
			"87654321": "http://example7.org",
		},
	}
	old = map[string]helpers.WebData{
		"cookie1": cookie1,
		"cookie2": cookie2,
		"cookie4": cookie4,
	}
	newcookie1 := helpers.WebData{
		Key: "secret-key1",
		Short: map[string]string{
			"abcdABCD": "http://example4.org",
		},
	}
	newcookie2 := helpers.WebData{
		Key: "",
		Short: map[string]string{
			"AbCdAbCd": "http://example5.org",
		},
	}
	newcookie3 := helpers.WebData{
		Key: "secret-key3",
		Short: map[string]string{
			"AbCdAbCd": "http://example6.org",
		},
	}
	newcookie5 := helpers.WebData{
		Key:   "secret-key5",
		Short: map[string]string{},
	}
	new = map[string]helpers.WebData{
		"cookie1": newcookie1,
		"cookie2": newcookie2,
		"cookie3": newcookie3,
		"cookie5": newcookie5,
	}
	result := make(map[string]helpers.WebData)
	rcookie1 := helpers.WebData{
		Key: "secret-key1",
		Short: map[string]string{
			"ABCDEFGH": "http://example1.org",
			"12345678": "http://example2.org",
			"abcdABCD": "http://example4.org",
		},
	}
	rcookie2 := helpers.WebData{
		Key: "secret-key2",
		Short: map[string]string{
			"87654321": "http://example3.org",
			"AbCdAbCd": "http://example5.org",
		},
	}
	rcookie3 := helpers.WebData{
		Key: "secret-key3",
		Short: map[string]string{
			"AbCdAbCd": "http://example6.org",
		},
	}
	rcookie4 := helpers.WebData{
		Key: "secret-key4",
		Short: map[string]string{
			"87654321": "http://example7.org",
		},
	}
	result = map[string]helpers.WebData{
		"cookie1": rcookie1,
		"cookie2": rcookie2,
		"cookie3": rcookie3,
		"cookie4": rcookie4,
		"cookie5": newcookie5,
	}
	data := mergeData(old, new)
	require.Equal(t, result, data)
}
