package storage

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"sync"

	"github.com/t1mon-ggg/go_shortner/internal/app/helpers"
	"github.com/t1mon-ggg/go_shortner/internal/app/models"
)

//FileDB - структура для работы с фаловым хранилищем данных
type FileDB struct {
	Name string      //имя файла
	file *os.File    //дескриптор для работы с файлом
	rw   *sync.Mutex //блокировка для защиты от одновременной записи
}

//NewFileDB - функция инициализирующая структура FileDB
func NewFileDB(name string) *FileDB {
	s := FileDB{}
	s.Name = name
	s.file = nil
	s.rw = &sync.Mutex{}
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
	f.rw.Lock()
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
	f.rw.Lock()
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
func (f *FileDB) Write(m models.ClientData) error {
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
	err = encoder.Encode(data)
	if err != nil {
		return err
	}
	f.Close()
	f.file = nil
	f.rw.Unlock()
	return nil
}

//ReadByCookie - чтение из файла
func (f *FileDB) readAllFile() ([]models.ClientData, error) {
	f.readFile()
	scanner := f.getScanner()
	m := make([]models.ClientData, 0)
	for scanner.Scan() {
		err := json.Unmarshal([]byte(scanner.Text()), &m)
		if err != nil {
			return nil, err
		}
	}
	f.Close()
	f.rw.Unlock()
	f.file = nil
	return m, nil
}

//TagByURL - поиск URL
func (f *FileDB) TagByURL(s, cookie string) (string, error) {
	data, err := f.readAllFile()
	if err != nil {
		return "", err
	}
	for _, value := range data {
		for _, url := range value.Short {
			if url.Long == s && value.Cookie == cookie {
				return url.Short, nil
			}
		}
	}
	return "", nil
}

//ReadByCookie - чтение из файла
func (f *FileDB) ReadByCookie(s string) (models.ClientData, error) {
	data, err := f.readAllFile()
	if err != nil {
		return models.ClientData{}, err
	}
	for _, value := range data {
		if value.Cookie == s {
			return value, nil
		}
	}
	f.Close()
	f.file = nil
	return models.ClientData{}, nil
}

//ReadByTag - чтение из файла
func (f *FileDB) ReadByTag(s string) (models.ShortData, error) {
	data, err := f.readAllFile()
	if err != nil {
		return models.ShortData{}, err
	}
	for _, cvalue := range data {
		for _, svalue := range cvalue.Short {
			if svalue.Short == s {
				return svalue, nil
			}
		}
	}
	return models.ShortData{}, nil
}

func (f *FileDB) deleteTag(task models.DelWorker) {
	data, err := f.readAllFile()
	if err != nil {
		log.Println("Error while reading file")
		return
	}
	for _, tag := range task.Tags {
		for i, user := range data {
			for j, url := range user.Short {
				if user.Cookie == task.Cookie && url.Short == tag {
					data[i].Short[j].Deleted = true
				}
			}
		}
	}
	f.rewriteFile()
	encoder := f.getCoder()
	err = encoder.Encode(data)
	if err != nil {
		log.Println("Error while rewriting file")
		return
	}
	f.Close()
	f.file = nil
	f.rw.Unlock()
}

func (f *FileDB) Cleaner(inputCh <-chan models.DelWorker, workers int) {
	fanOutChs := helpers.FanOut(inputCh, workers)
	for _, fanOutCh := range fanOutChs {
		go f.newWorker(fanOutCh)
	}
}

func (f *FileDB) newWorker(input <-chan models.DelWorker) {
	for task := range input {
		f.deleteTag(task)
	}
}
