package storage

import (
	"errors"

	"golang.org/x/exp/maps"

	"github.com/t1mon-ggg/go_shortner/internal/app/helpers"
)

type MemDB map[string]helpers.WebData

func NewMemDB() MemDB {
	var s MemDB
	s = make(map[string]helpers.WebData)
	return s
}

//Write - добавление данных в память
func (db MemDB) Write(m helpers.Data) error {
	mm := MemDB(m)
	var err = errors.New("DB not initialized")
	if db == nil {
		return err
	}
	err = errors.New("Invalid input data")
	if m == nil {
		return err
	}
	for i := range m {
		if !maps.Equal(db[i].Short, mm[i].Short) {
			maps.Copy(mm[i].Short, db[i].Short)
			e1 := mm[i]
			e2 := db[i]
			if e2.Key != "" {
				e1.Key = e2.Key
			}
			mm[i] = e1
		}
	}
	maps.Copy(db, mm)
	return nil
}

//ReadByCookie - чтение из памяти по cookie
func (db MemDB) ReadByCookie(s string) (helpers.Data, error) {
	var err = errors.New("DB not initialized")
	if db == nil {
		return nil, err
	}
	data := make(map[string]helpers.WebData)
	for cookie, webdata := range db {
		if cookie == s {
			data[s] = webdata
		}
	}
	return data, nil
}

//ReadByCookie - чтение из памяти по cookie
func (db MemDB) ReadByTag(s string) (map[string]string, error) {
	var err = errors.New("DB not initialized")
	if db == nil {
		return nil, err
	}
	data := make(map[string]string)
	for _, webdata := range db {
		for tag, url := range webdata.Short {
			if tag == s {
				data[tag] = url
			}
		}
	}
	return data, nil
}

//Close - освобождение области данных
func (db MemDB) Close() error {
	err := errors.New("DB not initialized")
	if db == nil {
		return err
	}
	for i := range db {
		delete(db, i)
	}
	return nil
}

//Ping - проверка наличия в памяти области данных
func (db MemDB) Ping() error {
	err := errors.New("DB not initialized")
	if db == nil {
		return err
	}
	return nil
}
