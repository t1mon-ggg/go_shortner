package storage

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
			defer f.Close()
			err := f.openFile(tt.args)
			require.NoError(t, err)
		})
	}
	t.Run("Remove test file", func(t *testing.T) {
		err := os.Remove("createme.txt")
		require.NoError(t, err)
	})
}

func TestFileDB_Write(t *testing.T) {
	tests := []struct {
		name string
		args map[string]string
	}{
		{
			name: "write json to file",
			args: map[string]string{
				"ABCDabcd": "https://yandex.ru",
			},
		},
		{
			name: "write  many jsons to file",
			args: map[string]string{
				"djsvndAD": "http://example.org",
				"12345678": "http://example1.org",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFileDB("createme.txt")
			defer f.Close()
			err := f.Write(tt.args)
			require.NoError(t, err)
		})
	}
	t.Run("Storage Close test", func(t *testing.T) {
		err := os.Remove("createme.txt")
		assert.Nil(t, err)
	})
}

func TestFileDB_Read(t *testing.T) {
	f := NewFileDB("createme.txt")
	data := map[string]string{
		"ABCDabcd": "https://yandex.ru",
		"djsvndAD": "http://example.org",
		"12345678": "http://example1.org",
	}
	f.Write(data)

	tests := []struct {
		name string
		f    *FileDB
		want map[string]string
	}{
		{
			name: "read json file",
			f:    f,
			want: map[string]string{
				"ABCDabcd": "https://yandex.ru",
				"djsvndAD": "http://example.org",
				"12345678": "http://example1.org",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.f.Read()
			defer tt.f.Close()
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
	t.Run("Storage Close test", func(t *testing.T) {
		err := os.Remove("createme.txt")
		assert.Nil(t, err)
	})
}
