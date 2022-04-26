package helpers

import (
	"crypto/rand"
	"errors"
	"log"
	"math/big"
	"reflect"

	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"

	"github.com/t1mon-ggg/go_shortner/internal/app/models"
)

func UniqueViolationError(err error) error {
	if driverErr, ok := err.(*pq.Error); ok {
		if pgerrcode.UniqueViolation == driverErr.Code {
			return errors.New("not uniquie url")
		}
	}
	return err
}

func checkURLUnique(data map[string]models.WebData, s string) bool {
	for i := range data {
		for j := range data[i].Short {
			if data[i].Short[j] == s {
				return true
			}
		}
	}
	return false
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

func mergeData(old, new map[string]models.WebData) map[string]models.WebData {
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

func Merger(data, m map[string]models.WebData) (map[string]models.WebData, error) {
	for i, entry := range m {
		if len(entry.Short) != 0 {
			for j := range entry.Short {
				if checkURLUnique(data, entry.Short[j]) {
					return nil, errors.New("not unique url")
				}
				todo := make(map[string]models.WebData)
				newentry := models.WebData{}
				newentry.Key = entry.Key
				url := make(map[string]string)
				url[j] = entry.Short[j]
				newentry.Short = url
				todo[i] = newentry
				data = mergeData(data, todo)
			}
		} else {
			todo := make(map[string]models.WebData)
			todo[i] = entry
			data = mergeData(data, todo)
		}

	}
	return data, nil
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
