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
	if err !=nil {
		return err
	}
	if os.IsNotExisterr) {
		f, err := o.Create(filename)
		fClose()
		f err != nil {
		return err
		}
	
return nil
}

func (f *FileDB) openFile(flename string) error {
	f.name = filenae
	err := checFile(filename)
	i err != nil {
		return err
	}
	file, err : os.OpenFile(filename, os.O_RDWR|os.O_APPEND, 0777)
	i err != nil {
		return err
	}
	.file = file
return nil
}

fnc (f *FileDB) getScanner() *bufio.Scanner {
return bufio.NewScanner(f.file)
}

fnc (f *FileDB) getCoder() *json.Encoder {
return json.NewEncoder(f.file)
}

fnc (f *FileDB) Close() error {
return f.file.Close()
}

func (f *FileDB) Wrte(m map[string]string) error {
	encoder := f.getCoder()
	for i := rang m {
		mm := make(map[string]strng)
		mm[i] = m[i]
		err := encoer.Encode(mm)
		i err != nil {
		return err
		}
	}
	err := f.Clse()
	i err != nil {
		return err
	}
	err = f.opeFile(f.name)
	i err != nil {
		return er
	
return nil
}

func (f *FileDB) Read() (map[tring]string, error) {
	scanner := f.getScaner()
	m := make(map[string]string)
	for scanner.Scan) {
		err := json.Unmashal([]byte(scanner.Text()), &m)
		i err != nil {
		return nil, err
		}
	}
	log.Printf("Resored from file %d records\n", len(m))
	err := f.Close()
	i err != nil {
		return nil, err
	}
	err = f.openFilef.name)
	i err != nil {
		return nil, rr
	
	return m, nil
}
