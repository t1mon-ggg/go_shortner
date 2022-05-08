package helpers

import (
	"crypto/rand"
	"errors"
	"log"
	"math/big"

	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"

	"github.com/t1mon-ggg/go_shortner/internal/app/models"
)

func UniqueViolationError(err error) bool {
	if driverErr, ok := err.(*pq.Error); ok {
		if pgerrcode.UniqueViolation == driverErr.Code {
			return true
		}
	}
	return false
}

func checkURLUnique(data []models.ClientData, cookie, url string) bool {
	for _, value := range data {
		for _, short := range value.Short {
			if short.Long == url && value.Cookie == cookie && !short.Deleted {
				return true
			}
		}
	}
	return false
}

func mergeURLs(old, new []models.ShortData) []models.ShortData {
	for _, newval := range new {
		count := 0
		for _, oldval := range old {
			if newval == oldval {
				count++
			}
		}
		if count == 0 {
			old = append(old, newval)
		}
	}
	return old
}

func mergeData(old []models.ClientData, new models.ClientData) []models.ClientData {
	if len(old) == 0 {
		old = append(old, new)
		return old
	}
	count := 0
	for i := range old {
		if old[i].Cookie == new.Cookie {
			count++
			old[i].Short = mergeURLs(old[i].Short, new.Short)
		}
	}
	if count == 0 {
		old = append(old, new)
	}
	return old
}

func Merger(old []models.ClientData, new models.ClientData) ([]models.ClientData, error) {
	for _, value := range new.Short {
		if checkURLUnique(old, new.Cookie, value.Long) {
			return old, errors.New("not unique url")
		}
	}
	old = mergeData(old, new)
	return old, nil
}

const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func RandStringRunes(n int) string {
	b := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			log.Fatal(err)
			return ""
		}
		b[i] = letters[num.Int64()]
	}

	return string(b)
}

func FanOut(inputCh <-chan models.DelWorker, workers int) []chan models.DelWorker {
	chs := make([]chan models.DelWorker, 0, workers)
	for i := 0; i < workers; i++ {
		ch := make(chan models.DelWorker)
		chs = append(chs, ch)
	}
	go func() {
		defer func(chs []chan models.DelWorker) {
			for _, ch := range chs {
				close(ch)
			}
		}(chs)

		for i := 0; ; i++ {
			if i == len(chs) {
				i = 0
			}

			delRequest, ok := <-inputCh
			if !ok {
				return
			}
			ch := chs[i]
			ch <- delRequest

		}
	}()
	return chs
}
