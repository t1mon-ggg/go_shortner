package storage

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/t1mon-ggg/go_shortner/internal/app/helpers"
)

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

func Test_FileDB_Write(t *testing.T) {
	tests := []struct {
		name string
		args helpers.Data
	}{
		{
			name: "write json to file",
			args: helpers.Data{
				"cookie1": {
					Key: "secret_key1",
					Short: map[string]string{
						"djsvndAD": "http://example.org",
						"12345678": "http://example1.org",
					},
				},
				"cookie2": {
					Key: "secret_key2",
					Short: map[string]string{
						"dsslkevn": "http://test.org",
						"12345678": "http://test1.org",
						"87654321": "http://test2.org",
					},
				},
				"cookie3": {
					Key: "secret_key2",
					Short: map[string]string{
						"wetgvsdc": "http://testing.org",
						"54756356": "http://testing1.org",
						"12353252": "http://testing2.org",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFileDB("createme.txt")
			err := f.Write(tt.args)
			require.NoError(t, err)
		})
	}
	t.Run("Storage Close test", func(t *testing.T) {
		err := os.Remove("createme.txt")
		require.NoError(t, err)
	})
}
