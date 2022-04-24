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
	var err error
	var f *os.File
	_, err = os.Stat(filename)
	if os.IsNotExist(err) {
		f, err = os.Create(filename)
		if err != nil {
			log.Println("1111")
			return err
		}
		f.Close()
	}
	return nil
}

func (f *FileDB) Ping() error {
	log.Println("Check connection to files storage")
	var err error
	f.file, err = os.OpenFile(f.Name, os.O_RDONLY, 0777)
	if err != nil {
		log.Println("File storage failed on opening file for read")
		return err
	}
	err = f.file.Close()
	if err != nil {
		log.Println("File storage failed on closing file after read")
		return err
	}
	f.file = nil
	f.file, err = os.OpenFile(f.Name, os.O_WRONLY, 0777)
	if err != nil {
		log.Println("File storage failed on opening file for write")
		return err
	}
	err = f.file.Close()
	if err != nil {
		log.Println("File storage failed on closing file after write")
		return err
	}
	f.file = nil
	log.Println("Connection to file storage confirmed")
	return nil
}

//readFile - создание файлового дескриптора для чтения из файла
func (f *FileDB) readFile() error {
	err := checkFile(f.Name)
	if err != nil {
		log.Println("2222")
		return err
	}
	file, err := os.OpenFile(f.Name, os.O_RDONLY, 0777)
	if err != nil {
		log.Println("3333")
		return err
	}
	f.file = file
	return nil
}

//rewriteFile - создание файлового дескриптора для перезаписи файла новыми данными
func (f *FileDB) rewriteFile() error {
	err := checkFile(f.Name)
	if err != nil {
		log.Println("4444")
		return err
	}
	file, err := os.OpenFile(f.Name, os.O_WRONLY|os.O_TRUNC, 0777)
	if err != nil {
		log.Println("5555")
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
	db, err := f.readAllFile()
	if err != nil {
		return err
	}
	db = mergeData(db, m)
	f.rewriteFile()
	encoder := f.getCoder()
	for i := range db {
		wr := make(helpers.Data)
		wr[i] = db[i]
		err := encoder.Encode(wr)
		if err != nil {
			return err
		}
	}
	f.Close()
	f.file = nil
	return nil
}

//ReadByCookie - чтение из файла
func (f *FileDB) readAllFile() (helpers.Data, error) {
	f.readFile()
	scanner := f.getScanner()
	m := make(helpers.Data)
	for scanner.Scan() {
		err := json.Unmarshal([]byte(scanner.Text()), &m)
		if err != nil {
			log.Println("9999")
			return nil, err
		}
	}
	f.Close()
	f.file = nil
	return m, nil
}

//ReadByCookie - чтение из файла
func (f *FileDB) ReadByCookie(s string) (helpers.Data, error) {
	f.readFile()
	scanner := f.getScanner()
	m := make(helpers.Data)
	for scanner.Scan() {
		err := json.Unmarshal([]byte(scanner.Text()), &m)
		if err != nil {
			log.Println("0000")
			return nil, err
		}
	}
	data := make(map[string]helpers.WebData)
	for cookie, webdata := range m {
		if cookie == s {
			data[s] = webdata
		}
	}
	f.Close()
	f.file = nil
	return data, nil
}

//ReadByTag - чтение из файла
func (f *FileDB) ReadByTag(s string) (map[string]string, error) {
	f.readFile()
	scanner := f.getScanner()
	m := make(helpers.Data)
	for scanner.Scan() {
		err := json.Unmarshal([]byte(scanner.Text()), &m)
		if err != nil {
			log.Println("!!!!")
			return nil, err
		}
	}
	f.Close()
	f.file = nil
	data := make(map[string]string)
	for _, webdata := range m {
		for tag, url := range webdata.Short {
			if tag == s {
				data[tag] = url
			}
		}
	}
	return data, nil
}
