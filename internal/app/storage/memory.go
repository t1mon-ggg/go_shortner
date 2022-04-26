package storage

import (
	"errors"

	"github.com/t1mon-ggg/go_shortner/internal/app/helpers"
	"github.com/t1mon-ggg/go_shortner/internal/app/models"
)

type MemDB map[string]models.WebData

func NewMemDB() MemDB {
	s := make(map[string]models.WebData)
	return s
}

//Write - добавление данных в память
func (data MemDB) Write(m map[string]models.WebData) error {
	err := errors.New("invalid input data")
	if len(m) == 0 {
		return err
	}
	data, err = helpers.Merger(data, m)
	if err != nil {
		return err
	}
	return nil
}

//TagByURL - чтение из памяти по cookie
func (data MemDB) TagByURL(s string) (string, error) {
	for i := range data {
		for j, url := range data[i].Short {
			if url == s {
				return j, nil
			}
		}
	}
	return "", nil
}

//ReadByCookie - чтение из памяти по cookie
func (data MemDB) ReadByCookie(s string) (map[string]models.WebData, error) {
	result := make(map[string]models.WebData)
	for cookie, webdata := range data {
		if cookie == s {
			result[s] = webdata
		}
	}
	return result, nil
}

//ReadByTag - чтение из памяти по cookie
func (data MemDB) ReadByTag(s string) (map[string]string, error) {
	result := make(map[string]string)
	for _, webdata := range data {
		for tag, url := range webdata.Short {
			if tag == s {
				result[tag] = url
			}
		}
	}
	return result, nil
}

//Close - освобождение области данных
func (data MemDB) Close() error {
	for i := range data {
		delete(data, i)
	}
	return nil
}

//Ping - проверка наличия в памяти области данных
func (data MemDB) Ping() error {
	return nil
}
