package storage

import (
	"errors"
	"reflect"

	"github.com/t1mon-ggg/go_shortner/internal/app/helpers"
)

type MemDB map[string]helpers.WebData

func NewMemDB() MemDB {
	var s MemDB
	s = make(map[string]helpers.WebData)
	return s
}

func mergeURLs(old, new map[string]string) map[string]string {
	if reflect.DeepEqual(old, new) {
		return old
	}
	for i := range new {
		if _, ok := old[i]; ok {
			if reflect.DeepEqual(old[i], new[i]) {
				continue
			}
		} else {
			old[i] = new[i]
		}
	}
	return old
}

func mergeData(old, new map[string]helpers.WebData) map[string]helpers.WebData {
	if reflect.DeepEqual(old, new) {
		return old
	}
	for i := range new {
		if _, ok := old[i]; ok {
			if reflect.DeepEqual(old[i], new[i]) {
				continue
			} else {
				entry := old[i]
				newentry := new[i]
				if newentry.Key != "" && newentry.Key != entry.Key {
					entry.Key = newentry.Key
				}
				entry.Short = mergeURLs(entry.Short, newentry.Short)
				old[i] = entry
			}
		} else {
			old[i] = new[i]
		}
	}
	return old
}

//Write - добавление данных в память
func (db MemDB) Write(m helpers.Data) error {
	var err = errors.New("DB not initialized")
	if db == nil {
		return err
	}
	err = errors.New("Invalid input data")
	if m == nil {
		return err
	}
	db = mergeData(db, m)
	return nil
}

//ReadByCookie - чтение из памяти по cookie
func (db MemDB) TagByURL(s string) (string, error) {
	for i := range db {
		for j, url := range db[i].Short {
			if url == s {
				return j, nil
			}
		}
	}
	return "", nil
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
