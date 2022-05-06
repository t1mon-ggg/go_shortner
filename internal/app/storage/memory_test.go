package storage

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/t1mon-ggg/go_shortner/internal/app/helpers"
	"github.com/t1mon-ggg/go_shortner/internal/app/models"
)

func (data *ram) testPrepare(t *testing.T) {
	d := []models.ClientData{
		{
			Cookie: "cookie1",
			Key:    "secret_key1",
			Short: []models.ShortData{
				{
					Short:   "abcdABC1",
					Long:    "http://example1.org",
					Deleted: false,
				},
			},
		},
		{
			Cookie: "cookie2",
			Key:    "secret_key2",
			Short: []models.ShortData{
				{
					Short:   "abcdABC2",
					Long:    "http://example2.org",
					Deleted: false,
				},
			},
		},
		{
			Cookie: "cookie3",
			Key:    "secret_key3",
			Short: []models.ShortData{
				{
					Short:   "abcdABC3",
					Long:    "http://example3.org",
					Deleted: false,
				},
			},
		},
	}
	for _, value := range d {
		err := data.Write(value)
		require.NoError(t, err)
	}
}

func Test_MEM_Write(t *testing.T) {
	db := newRAM()
	data := models.ClientData{
		Cookie: "cookie1",
		Key:    "secret_key",
		Short: []models.ShortData{
			{
				Short: "Short1",
				Long:  "Long1",
			},
		},
	}
	exp := newRAM()
	exp.DB, _ = helpers.Merger(exp.DB, data)
	err := db.Write(data)
	require.NoError(t, err)
	require.Equal(t, exp, db)
}

func Test_MEM_ReadByCookie(t *testing.T) {
	db := newRAM()
	db.testPrepare(t)
	exp := models.ClientData{
		Cookie: "cookie2",
		Key:    "secret_key2",
		Short: []models.ShortData{
			{
				Short: "abcdABC2",
				Long:  "http://example2.org",
			},
		},
	}
	val, err := db.ReadByCookie("cookie2")
	require.NoError(t, err)
	require.Equal(t, exp, val)
}

func Test_MEM_ReadByTag(t *testing.T) {
	db := newRAM()
	db.testPrepare(t)
	expected := models.ShortData{Short: "abcdABC2", Long: "http://example2.org"}
	val, err := db.ReadByTag("abcdABC2")
	require.NoError(t, err)
	require.Equal(t, expected, val)
}

func Test_MEM_Close(t *testing.T) {
	db := newRAM()
	db.testPrepare(t)
	err := db.Close()
	require.NoError(t, err)
	require.Equal(t, ram{}, *db)
}

func Test_MEM_Ping(t *testing.T) {
	db := newRAM()
	db.testPrepare(t)
	err := db.Ping()
	require.NoError(t, err)
}

func Test_MEM_Delete(t *testing.T) {
	r := models.ClientData{
		Cookie: "cookie2",
		Key:    "secret_key2",
		Short: []models.ShortData{
			{
				Short:   "abcdABC2",
				Long:    "http://example2.org",
				Deleted: true,
			},
		},
	}
	db := newRAM()
	db.testPrepare(t)
	task := models.DelWorker{
		Cookie: "cookie2",
		Tags:   []string{"abcdABC2"},
	}
	db.deleteTag(task)
	time.Sleep(5 * time.Second)
	d, err := db.ReadByCookie("cookie2")
	require.NoError(t, err)
	require.Equal(t, r, d)

}
