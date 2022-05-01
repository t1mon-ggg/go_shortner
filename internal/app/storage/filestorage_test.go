package storage

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/t1mon-ggg/go_shortner/internal/app/models"
)

//Ping() error
func Test_File_Ping(t *testing.T) {
	var err error
	f := FileDB{}
	f.Name = "createme.txt"
	f.file, err = os.Create("createme.txt")
	require.NoError(t, err)
	f.Close()
	f.file = nil
	err = f.Ping()
	require.NoError(t, err)
	err = os.Remove(f.Name)
	require.NoError(t, err)

}
func Test_openFile(t *testing.T) {
	type args struct {
		filename string
	}
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
			f := FileDB{}
			f.Name = tt.args
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

//Write(models.ClientData) error
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
	f := NewFileDB("createme.txt")
	for _, value := range data {
		err := f.Write(value)
		require.NoError(t, err)
	}
	err := os.Remove("createme.txt")
	require.NoError(t, err)
}

func (f *FileDB) test_prepare(t *testing.T) {
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

//ReadByCookie(string) (models.ClientData, error)
func Test_FileDB_ReadByCookie(t *testing.T) {
	f := NewFileDB("createme.txt")
	f.test_prepare(t)
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

//ReadByTag(string) (models.ShortData, error)
func Test_FileDB_ReadByTag(t *testing.T) {
	f := NewFileDB("createme.txt")
	f.test_prepare(t)
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

//TagByURL(string) (string, error)
func Test_FileDB_TagByURL(t *testing.T) {
	f := NewFileDB("createme.txt")
	f.test_prepare(t)
	exp := "abcdABC2"
	data, err := f.TagByURL("http://example2.org")
	require.NoError(t, err)
	require.Equal(t, exp, data)
	err = os.Remove("createme.txt")
	require.NoError(t, err)

}
