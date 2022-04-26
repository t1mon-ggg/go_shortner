package storage

import (
	"bufio"
	"encoding/json"
	"log"
	"os"

	"github.com/t1mon-ggg/go_shortner/internal/app/helpers"
	"github.com/t1mon-ggg/go_shortner/internal/app/models"
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
func (f *FileDB) Write(m map[string]models.WebData) error {
	data, err := f.readAllFile()
	if err != nil {
		return err
	}
	data, err = helpers.Merger(data, m)
	if err != nil {
		return err
	}
	f.rewriteFile()
	encoder := f.getCoder()
	for i := range data {
		wr := make(map[string]models.WebData)
		wr[i] = data[i]
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
func (f *FileDB) readAllFile() (map[string]models.WebData, error) {
	f.readFile()
	scanner := f.getScanner()
	m := make(map[string]models.WebData)
	for scanner.Scan() {
		err := json.Unmarshal([]byte(scanner.Text()), &m)
		if err != nil {
			return nil, err
		}
	}
	f.Close()
	f.file = nil
	return m, nil
}

//TagByURL - поиск URL
func (f *FileDB) TagByURL(s string) (string, error) {
	data, err := f.readAllFile()
	if err != nil {
		return "", err
	}
	for i := range data {
		for j, url := range data[i].Short {
			if url == s {
				return j, nil
			}
		}
	}
	return "", nil
}

//ReadByCookie - чтение из файла
func (f *FileDB) ReadByCookie(s string) (map[string]models.WebData, error) {
	f.readFile()
	scanner := f.getScanner()
	m := make(map[string]models.WebData)
	for scanner.Scan() {
		err := json.Unmarshal([]byte(scanner.Text()), &m)
		if err != nil {
			return nil, err
		}
	}
	data := make(map[string]models.WebData)
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
	m := make(map[string]models.WebData)
	for scanner.Scan() {
		err := json.Unmarshal([]byte(scanner.Text()), &m)
		if err != nil {
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
