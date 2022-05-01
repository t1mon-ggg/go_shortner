package storage

import (
	"github.com/t1mon-ggg/go_shortner/internal/app/helpers"
	"github.com/t1mon-ggg/go_shortner/internal/app/models"
)

type MemDB []models.ClientData

//NewMemDB - new in memory storage
func NewMemDB() *MemDB {
	s := MemDB(make([]models.ClientData, 0))
	return &s
}

func (data *MemDB) clientExist(m models.ClientData) bool {
	for _, value := range *data {
		if value.Cookie == m.Cookie {
			return true
		}
	}
	return false
}

//Write - добавление данных в память
func (data *MemDB) Write(m models.ClientData) error {
	newData, err := helpers.Merger(*data, m)
	if err != nil {
		return err
	}
	*data = newData
	return nil
}

//TagByURL - чтение из памяти по cookie
func (data *MemDB) TagByURL(s string) (string, error) {
	for _, value := range *data {
		for _, url := range value.Short {
			if url.Long == s {
				return url.Short, nil
			}
		}
	}
	return "", nil
}

//ReadByCookie - чтение из памяти по cookie
func (data *MemDB) ReadByCookie(s string) (models.ClientData, error) {
	for _, value := range *data {
		if value.Cookie == s {
			return value, nil
		}
	}
	return models.ClientData{}, nil
}

//ReadByTag - чтение из памяти по cookie
func (data *MemDB) ReadByTag(s string) (models.ShortData, error) {
	for _, userValue := range *data {
		for _, urlValue := range userValue.Short {
			if urlValue.Short == s {
				return urlValue, nil
			}
		}
	}
	return models.ShortData{}, nil
}

//Close - освобождение области данных
func (data *MemDB) Close() error {
	*data = MemDB{}
	return nil
}

//Ping - проверка наличия в памяти области данных
func (data MemDB) Ping() error {
	return nil
}
