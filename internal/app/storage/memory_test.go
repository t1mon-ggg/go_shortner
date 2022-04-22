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
