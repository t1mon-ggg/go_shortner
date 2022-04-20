package storage

import (
	"bufio"
	"encoding/json"
	"log"
	"os"

	"github.com/t1mon-ggg/go_shortner/internal/app/helpers"
)

//FileDB - структура для работы с фаловым хранилищем данных
type FileDB struct {
	Name string   //имя файла
	file *os.File //дескриптор для работы с файлом
}

//NewFileDB - функция инициализирующая структура FileDB
func NewFileDB(name string) *FileDB {
	s := FileDB{}
	s.Name = name
	s.file = nil
	return &s
}

//checkFile - функция проверки существования файла и его создания
func checkFile(filename string) error {
	_, staterr := os.Stat(filename)
	if os.IsNotExist(staterr) {
		f, createerr := os.Create(filename)
		if createerr != nil {
			return createerr
		}
		f.Close()
	}
	return nil
}

func (f *FileDB) Ping() error {
	var err error
	f.file, err = os.OpenFile(f.Name, os.O_RDONLY, 0777)
	if err != nil {
		return err
	}
	err = f.file.Close()
	if err != nil {
		return err
	}
	f.file = nil
	f.file, err = os.OpenFile(f.Name, os.O_WRONLY, 0777)
	if err != nil {
		return err
	}
	err = f.file.Close()
	if err != nil {
		return err
	}
	f.file = nil
	return nil
}

//readFile - создание файлового дескриптора для чтения из файла
func (f *FileDB) readFile() error {
	err := checkFile(f.Name)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(f.Name, os.O_RDONLY, 0777)
	if err != nil {
		return err
	}
	f.file = file
	return nil
}

//rewriteFile - создание файлового дескриптора для перезаписи файла новыми данными
func (f *FileDB) rewriteFile() error {
	err := checkFile(f.Name)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(f.Name, os.O_WRONLY|os.O_TRUNC, 0777)
	if err != nil {
		return err
	}
	f.file = file
	return nil
}

//getScanner - создание дескритрора для потокового чтения json из файла
func (f *FileDB) getScanner() *bufio.Scanner {
	return bufio.NewScanner(f.file)
}

//getCoder - создание дескритрора для потоковой записи json в файл
func (f *FileDB) getCoder() *json.Encoder {
	return json.NewEncoder(f.file)
}

//Close - закрытие файлового дескриптора после операцияй чтения/записи файла
func (f *FileDB) Close() error {
	return f.file.Close()
}

//Write - запись в файл
func (f *FileDB) Write(m helpers.Data) error {
	f.rewriteFile()
	encoder := f.getCoder()
	for i := range m {
		mm := make(helpers.Data)
		mm[i] = m[i]
		err := encoder.Encode(mm)
		if err != nil {
			return err
		}
	}
	f.Close()
	f.file = nil
	return nil
}

//Read - чтение из файла
func (f *FileDB) Read() (helpers.Data, error) {
	f.readFile()
	scanner := f.getScanner()
	m := make(helpers.Data)
	for scanner.Scan() {
		err := json.Unmarshal([]byte(scanner.Text()), &m)
		if err != nil {
			return nil, err
		}
	}
	log.Printf("Restored from file %d records\n", len(m))
	f.Close()
	f.file = nil
	return m, nil
}
