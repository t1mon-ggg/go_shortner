package storage

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewCoder(t *testing.T) {
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := FileDB{}
			err := f.NewCoder(tt.args)
			require.NoError(t, err)
		})
	}
}

func TestFileDB_Write(t *testing.T) {
	f := FileDB{}
	_ = f.NewCoder("createme.txt")
	tests := []struct {
		name string
		f    *FileDB
		args map[string]string
	}{
		{
			name: "write json to file",
			f:    &f,
			args: map[string]string{
				"ABCDabcd": "https://yandex.ru",
			},
		},
		{
			name: "write  many jsons to file",
			f:    &f,
			args: map[string]string{
				"djsvndAD": "http://example.org",
				"12345678": "http://example1.org",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.f.Write(tt.args)
			require.NoError(t, err)
		})
	}
}

func TestFileDB_Read(t *testing.T) {
	f := FileDB{}
	_ = f.NewCoder("createme.txt")
	tests := []struct {
		name string
		f    *FileDB
		want map[string]string
	}{
		{
			name: "read json file",
			f:    &f,
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
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
