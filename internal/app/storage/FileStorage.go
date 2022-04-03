package storage

import (
	"bufio"
	"encoding/json"
	"os"
)

type FileDB struct {
	file    *os.File
	decoder *bufio.Scanner
	encoder *json.Encoder
}

func checkFile(filename string) error {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		_, err := os.Create(filename)
		if err != nil {
			return err
		}
	}
	return nil
}

func NewCoder(filename string) (*FileDB, error) {
	err := checkFile(filename)
	if err != nil {
		return nil, err
	}
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_APPEND, 0777)
	return &FileDB{
			file:    file,
			decoder: bufio.NewScanner(file),
			encoder: json.NewEncoder(file),
		},
		nil
}

func (f *FileDB) Close() error {
	return f.file.Close()
}

func (f *FileDB) Write(m map[string]string) error {
	for i := range m {
		mm := make(map[string]string)
		mm[i] = m[i]
		err := f.encoder.Encode(mm)
		if err != nil {
			return err
		}
	}

	return nil
}

func (f *FileDB) Read() (map[string]string, error) {
	m := make(map[string]string)
	for f.decoder.Scan() {
		err := json.Unmarshal([]byte(f.decoder.Text()), &m)
		if err != nil {
			return nil, err
		}
	}
	return m, nil
}
