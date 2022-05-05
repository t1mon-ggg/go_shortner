package storage

import (
	"sync"

	"github.com/t1mon-ggg/go_shortner/internal/app/helpers"
	"github.com/t1mon-ggg/go_shortner/internal/app/models"
)

type MemDB struct {
	DB  []models.ClientData
	Mux *sync.RWMutex
}

//NewMemDB - new in memory storage
func NewMemDB() *MemDB {
	s := MemDB{}
	s.DB = make([]models.ClientData, 0)
	s.Mux = &sync.RWMutex{}
	return &s
}

func (data *MemDB) clientExist(m models.ClientData) bool {
	(*data).Mux.RLock()
	for _, value := range (*data).DB {
		if value.Cookie == m.Cookie {
			(*data).Mux.RUnlock()
			return true
		}
	}
	(*data).Mux.RUnlock()
	return false
}

//Write - добавление данных в память
func (data *MemDB) Write(m models.ClientData) error {
	(*data).Mux.Lock()
	newData, err := helpers.Merger((*data).DB, m)
	if err != nil {
		(*data).Mux.Unlock()
		return err
	}
	(*data).DB = newData
	(*data).Mux.Unlock()
	return nil
}

//TagByURL - чтение из памяти по cookie
func (data *MemDB) TagByURL(s, cookie string) (string, error) {
	(*data).Mux.RLock()
	for _, value := range (*data).DB {
		for _, url := range value.Short {
			if url.Long == s && value.Cookie == cookie {
				(*data).Mux.RUnlock()
				return url.Short, nil
			}
		}
	}
	(*data).Mux.RUnlock()
	return "", nil
}

//ReadByCookie - чтение из памяти по cookie
func (data *MemDB) ReadByCookie(s string) (models.ClientData, error) {
	(*data).Mux.RLock()
	for _, value := range (*data).DB {
		if value.Cookie == s {
			(*data).Mux.RUnlock()
			return value, nil
		}
	}
	(*data).Mux.RUnlock()
	return models.ClientData{}, nil
}

//ReadByTag - чтение из памяти по cookie
func (data *MemDB) ReadByTag(s string) (models.ShortData, error) {
	(*data).Mux.RLock()
	for _, userValue := range (*data).DB {
		for _, urlValue := range userValue.Short {
			if urlValue.Short == s {
				(*data).Mux.RUnlock()
				return urlValue, nil
			}
		}
	}
	(*data).Mux.RUnlock()
	return models.ShortData{}, nil
}

//Close - освобождение области данных
func (data *MemDB) Close() error {
	(*data).Mux.Lock()
	*data = MemDB{}
	(*data).Mux.Unlock()
	return nil
}

//Ping - проверка наличия в памяти области данных
func (data MemDB) Ping() error {
	return nil
}

func (data *MemDB) Cleaner(inputCh <-chan models.DelWorker, workers int) {
	fanOutChs := helpers.FanOut(inputCh, workers)
	for _, fanOutCh := range fanOutChs {
		go data.newWorker(fanOutCh)
	}
}

func (data *MemDB) deleteTag(task models.DelWorker) {
	(*data).Mux.Lock()
	for _, tag := range task.Tags {
		for i, user := range (*data).DB {
			for j, url := range user.Short {
				if user.Cookie == task.Cookie && url.Short == tag {
					(*data).DB[i].Short[j].Deleted = true
				}
			}
		}
	}
	(*data).Mux.Unlock()

}

func (data *MemDB) newWorker(input <-chan models.DelWorker) {
	for task := range input {
		data.deleteTag(task)
	}
}
