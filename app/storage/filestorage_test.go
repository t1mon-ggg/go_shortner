package storage

import (
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/t1mon-ggg/go_shortner/app/models"
)

// Ping() error
func Test_File_Ping(t *testing.T) {
	f := NewFile("createme.txt")
	err := checkFile(f.name)
	require.NoError(t, err)
	err = f.Ping()
	require.NoError(t, err)
	err = os.Remove(f.name)
	require.NoError(t, err)

}
func Test_openFile(t *testing.T) {
	tests := []struct {
		name string
		args string
	}{
		{
			name: "NotExistent file",
			args: "createme.txt",
		},
		{
			name: "Duplicated file",
			args: "createme.txt",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := fileStorage{}
			f.rw = &sync.Mutex{}
			f.name = tt.args
			defer f.Close()
			err := f.readFile()
			require.NoError(t, err)
		})
	}
	t.Run("Remove test file", func(t *testing.T) {
		err := os.Remove("createme.txt")
		require.NoError(t, err)
	})
}

// Write(models.ClientData) error
func Test_FileDB_Write(t *testing.T) {
	data := []models.ClientData{
		{
			Cookie: "cookie1",
			Key:    "secret_key1",
			Short: []models.ShortData{
				{
					Short: "abcdABC1",
					Long:  "http://example1.org",
				},
			},
		},
		{
			Cookie: "cookie2",
			Key:    "secret_key2",
			Short: []models.ShortData{
				{
					Short: "abcdABC2",
					Long:  "http://example2.org",
				},
			},
		},
		{
			Cookie: "cookie3",
			Key:    "secret_key3",
			Short: []models.ShortData{
				{
					Short: "abcdABC3",
					Long:  "http://example3.org",
				},
			},
		},
	}
	f := NewFile("createme.txt")
	for _, value := range data {
		err := f.Write(value)
		require.NoError(t, err)
	}
	err := os.Remove("createme.txt")
	require.NoError(t, err)
}

func (f *fileStorage) testPrepare(t *testing.T) {
	data := []models.ClientData{
		{
			Cookie: "cookie1",
			Key:    "secret_key1",
			Short: []models.ShortData{
				{
					Short: "abcdABC1",
					Long:  "http://example1.org",
				},
			},
		},
		{
			Cookie: "cookie2",
			Key:    "secret_key2",
			Short: []models.ShortData{
				{
					Short: "abcdABC2",
					Long:  "http://example2.org",
				},
			},
		},
		{
			Cookie: "cookie3",
			Key:    "secret_key3",
			Short: []models.ShortData{
				{
					Short: "abcdABC3",
					Long:  "http://example3.org",
				},
			},
		},
	}
	for _, value := range data {
		err := f.Write(value)
		require.NoError(t, err)
	}
}

// ReadByCookie(string) (models.ClientData, error)
func Test_FileDB_ReadByCookie(t *testing.T) {
	f := NewFile("createme.txt")
	f.testPrepare(t)
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
	data, err := f.ReadByCookie("cookie2")
	require.NoError(t, err)
	require.Equal(t, exp, data)
	err = os.Remove("createme.txt")
	require.NoError(t, err)
}

// ReadByTag(string) (models.ShortData, error)
func Test_FileDB_ReadByTag(t *testing.T) {
	f := NewFile("createme.txt")
	f.testPrepare(t)
	exp := models.ShortData{
		Short: "abcdABC2",
		Long:  "http://example2.org",
	}
	data, err := f.ReadByTag("abcdABC2")
	require.NoError(t, err)
	require.Equal(t, exp, data)
	err = os.Remove("createme.txt")
	require.NoError(t, err)
}

// TagByURL(string) (string, error)
func Test_FileDB_TagByURL(t *testing.T) {
	f := NewFile("createme.txt")
	f.testPrepare(t)
	exp := "abcdABC2"
	data, err := f.TagByURL("http://example2.org", "cookie2")
	require.NoError(t, err)
	require.Equal(t, exp, data)
	err = os.Remove("createme.txt")
	require.NoError(t, err)

}

// Delete([]string) error
func Test_FileDB_Delete(t *testing.T) {
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
	f := NewFile("createme.txt")
	f.testPrepare(t)
	task := models.DelWorker{Cookie: "cookie2", Tags: []string{"abcdABC2"}}
	f.deleteTag(task)
	time.Sleep(5 * time.Second)
	d, err := f.ReadByCookie("cookie2")
	require.NoError(t, err)
	require.Equal(t, r, d)
	err = os.Remove("createme.txt")
	require.NoError(t, err)
}

func Test_FileDB_GetStats(t *testing.T) {
	f := NewFile("createme.txt")
	f.testPrepare(t)
	val, err := f.GetStats()
	require.NoError(t, err)
	require.NotEmpty(t, val)
	err = os.Remove("createme.txt")
	require.NoError(t, err)
}

func Test_FileDB_Cleaner(t *testing.T) {
	db := NewFile("createme.txt")
	db.testPrepare(t)
	type args struct {
		done    <-chan struct{}
		wg      *sync.WaitGroup
		inputCh chan models.DelWorker
		workers int
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "start cleaner",
			args: args{
				done:    make(<-chan struct{}),
				wg:      &sync.WaitGroup{},
				inputCh: make(chan models.DelWorker),
				workers: 10,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db.Cleaner(tt.args.done, tt.args.wg, tt.args.inputCh, tt.args.workers)
			tt.args.inputCh <- models.DelWorker{Cookie: "cookie3", Tags: []string{"abcdABC3"}}
			time.Sleep(5 * time.Second)
			data, _ := db.ReadByTag("abcdABC3")
			require.True(t, data.Deleted)
		})
	}
	err := os.Remove("createme.txt")
	require.NoError(t, err)
}
