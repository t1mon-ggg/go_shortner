package storage

import (
	"os"
	"sync"

	"github.com/t1mon-ggg/go_shortner/app/helpers"
	"github.com/t1mon-ggg/go_shortner/app/models"
)

type ram struct {
	Mux *sync.RWMutex
	DB  []models.ClientData
}

// Newram - new in memory storage
func NewRAM() *ram {
	s := ram{}
	s.DB = make([]models.ClientData, 0)
	s.Mux = &sync.RWMutex{}
	return &s
}

// Write - добавление данных в память
func (data *ram) Write(m models.ClientData) error {
	data.Mux.Lock()
	newData, err := helpers.Merger(data.DB, m)
	if err != nil {
		data.Mux.Unlock()
		return err
	}
	data.DB = newData
	data.Mux.Unlock()
	return nil
}

// TagByURL - чтение из памяти по cookie
func (data *ram) TagByURL(s, cookie string) (string, error) {
	data.Mux.RLock()
	for _, value := range data.DB {
		for _, url := range value.Short {
			if url.Long == s && value.Cookie == cookie {
				data.Mux.RUnlock()
				return url.Short, nil
			}
		}
	}
	data.Mux.RUnlock()
	return "", nil
}

// ReadByCookie - чтение из памяти по cookie
func (data *ram) ReadByCookie(s string) (models.ClientData, error) {
	data.Mux.RLock()
	for _, value := range data.DB {
		if value.Cookie == s {
			data.Mux.RUnlock()
			return value, nil
		}
	}
	data.Mux.RUnlock()
	return models.ClientData{}, nil
}

// ReadByTag - чтение из памяти по cookie
func (data *ram) ReadByTag(s string) (models.ShortData, error) {
	data.Mux.RLock()
	for _, userValue := range data.DB {
		for _, urlValue := range userValue.Short {
			if urlValue.Short == s {
				data.Mux.RUnlock()
				return urlValue, nil
			}
		}
	}
	data.Mux.RUnlock()
	return models.ShortData{}, nil
}

// Close - освобождение области данных
func (data *ram) Close() error {
	data.Mux.Lock()
	*data = ram{}
	return nil
}

// Ping - проверка наличия в памяти области данных
func (data ram) Ping() error {
	return nil
}

// Cleaner - delete task worker creator
func (data *ram) Cleaner(done <-chan os.Signal, wg *sync.WaitGroup, inputCh <-chan models.DelWorker, workers int) {
	fanOutChs := helpers.FanOut(wg, inputCh, workers)
	for _, fanOutCh := range fanOutChs {
		go data.newWorker(done, wg, fanOutCh)
	}
}

// deleteTag - mark tag as deleted
func (data *ram) deleteTag(task models.DelWorker) {
	data.Mux.Lock()
	for _, tag := range task.Tags {
		for i, user := range data.DB {
			for j, url := range user.Short {
				if user.Cookie == task.Cookie && url.Short == tag {
					data.DB[i].Short[j].Deleted = true
				}
			}
		}
	}
	data.Mux.Unlock()

}

// newWorker - delete task worker
func (data *ram) newWorker(done <-chan os.Signal, wg *sync.WaitGroup, input <-chan models.DelWorker) {
	for {
		select {
		case task := <-input:
			data.deleteTag(task)
		case <-done:
			wg.Done()
			return
		}
	}
}
