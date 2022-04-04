package storage

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
)

type FileDB struct {
	name string
	file *os.File
}

func NewFileDB(name string) *FileDB {
	s := &FileDB{}
	s.openFile(name)
	return s
}

type DB map[string]string

func checkFile(filename string) error {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		f, err := os.Create(filename)
		f.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func (f *FileDB) openFile(filename string) error {
	f.name = filename
	err := checkFile(filename)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_APPEND, 0777)
	if err != nil {
		return err
	}
	f.file = file
	return nil
}

func (f *FileDB) getScanner() *bufio.Scanner {
	return bufio.NewScanner(f.file)
}

func (f *FileDB) getCoder() *json.Encoder {
	return json.NewEncoder(f.file)
}

func (f *FileDB) Close() error {
	return f.file.Close()
}

func (f *FileDB) Write(m map[string]string) error {
	encoder := f.getCoder()
	for i := range m {
		mm := make(map[string]string)
		mm[i] = m[i]
		err := encoder.Encode(mm)
		if err != nil {
			return err
		}
	}
	err := f.Close()
	if err != nil {
		return err
	}
	err = f.openFile(f.name)
	if err != nil {
		return err
	}
	return nil
}

func (f *FileDB) Read() (map[string]string, error) {
	scanner := f.getScanner()
	m := make(map[string]string)
	for scanner.Scan() {
		err := json.Unmarshal([]byte(scanner.Text()), &m)
		if err != nil {
			return nil, err
		}
	}
	log.Printf("Restored from file %d records\n", len(m))
	err := f.Close()
	if err != nil {
		return nil, err
	}
	err = f.openFile(f.name)
	if err != nil {
		return nil, err
	}
	return m, nil
}
