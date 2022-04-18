package webhandlers

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi"

	"github.com/t1mon-ggg/go_shortner/internal/app/config"
	"github.com/t1mon-ggg/go_shortner/internal/app/helpers"
	"github.com/t1mon-ggg/go_shortner/internal/app/rand"
	"github.com/t1mon-ggg/go_shortner/internal/app/storage"
)

type app struct {
	Storage storage.FileDB
	Config  config.OsVars
	Data    helpers.Data
}

func NewData() helpers.Data {
	s := make(helpers.Data)
	return s
}

func NewApp() *app {
	s := app{}
	s.Storage = storage.FileDB{}
	s.Config = config.OsVars{}
	s.Data = NewData()
	return &s
}

func (db *app) Router(r chi.Router) {
	r.Get("/", defaultGetHandler)
	r.Get("/{^[a-zA-Z]}", db.getHandler)
	r.Get("/api/user/urls", db.userURLs)
	r.Post("/", db.postHandler)
	r.Post("/api/shorten", db.postAPIHandler)
	r.MethodNotAllowed(otherHandler)
}

func defaultGetHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Empty request", http.StatusBadRequest)
}

func otherHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Bad request", http.StatusBadRequest)
}

func (db *app) userURLs(w http.ResponseWriter, r *http.Request) {
	type answer struct {
		Short    string `json:"short_url"`
		Original string `json:"original_url"`
	}
	var newcookieval string
	cookie, err := r.Cookie("Client_ID")
	if err != nil {
		str := strings.Split(w.Header().Get("Set-Cookie"), "=")
		newcookieval = str[1]
	}
	var value string
	if len(newcookieval) != 0 {
		value = newcookieval[:32]
	} else {
		value = cookie.Value[:32]
	}
	if data, ok := db.Data[value]; ok {
		if len(data.Short) == 0 {
			http.Error(w, "No Content", http.StatusNoContent)
			return
		}
		a := make([]answer, 0)
		for i := range data.Short {
			a = append(a, answer{Short: fmt.Sprintf("%s/%s", db.Config.BaseURL, i), Original: data.Short[i]})
		}
		d, err := json.MarshalIndent(a, "", "\t")
		if err != nil {
			http.Error(w, "Json Error", http.StatusInternalServerError)
			return
		}
		w.Header().Add("Content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(d)
	} else {
		http.Error(w, "No Content", http.StatusNoContent)
		return
	}

}

func (db *app) postHandler(w http.ResponseWriter, r *http.Request) {
	var newcookieval string
	cookie, err := r.Cookie("Client_ID")
	if err != nil {
		str := strings.Split(w.Header().Get("Set-Cookie"), "=")
		newcookieval = str[1]
	}
	var value string
	if len(newcookieval) != 0 {
		value = newcookieval[:32]
	} else {
		value = cookie.Value[:32]
	}
	defer r.Body.Close()
	blongURL, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
	}
	slongURL := string(blongURL)
	surl := rand.RandStringRunes(8)
	if entry, ok := db.Data[value]; ok {
		entry.Short[surl] = slongURL
		db.Data[value] = entry
	}
	db.Storage.Write(db.Data)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf("%s/%s", db.Config.BaseURL, surl)))
}

func (db *app) postAPIHandler(w http.ResponseWriter, r *http.Request) {
	var newcookieval string
	cookie, err := r.Cookie("Client_ID")
	if err != nil {
		str := strings.Split(w.Header().Get("Set-Cookie"), "=")
		newcookieval = str[1]
	}
	var value string
	if len(newcookieval) != 0 {
		value = newcookieval[:32]
	} else {
		value = cookie.Value[:32]
	}
	type sURL struct {
		ShortURL string `json:"result"`
	}
	type lURL struct {
		LongURL string `json:"url"`
	}
	defer r.Body.Close()
	ctype := r.Header.Get("Content-Type")
	if ctype != "application/json" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	longURL := lURL{}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("Body Error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	err = json.Unmarshal(body, &longURL)
	if err != nil {
		log.Println("JSON Unmarshal error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	short := rand.RandStringRunes(8)
	if entry, ok := db.Data[value]; ok {
		entry.Short[short] = longURL.LongURL
		db.Data[value] = entry
	}
	db.Storage.Write(db.Data)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	jbody := sURL{ShortURL: fmt.Sprintf("%s/%s", (*db).Config.BaseURL, short)}
	abody, err := json.Marshal(jbody)
	if err != nil {
		log.Println("JSON Marshal error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Write(abody)
}

func (db app) getHandler(w http.ResponseWriter, r *http.Request) {
	p := r.RequestURI
	p = p[1:]
	if len(p) != 8 {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	var count int
	for _, entry := range db.Data {
		if url, ok := entry.Short[p]; ok {
			w.Header().Set("Location", url)
			w.WriteHeader(http.StatusTemporaryRedirect)
			w.Write([]byte{})
			count++
		}
	}
	if count == 0 {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

}

func DecompressRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if (strings.Contains(r.Header.Get("Content-Encoding"), "gzip")) || (strings.Contains(r.Header.Get("Content-Encoding"), "br")) || (strings.Contains(r.Header.Get("Content-Encoding"), "deflate")) {
			defer r.Body.Close()
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			defer gz.Close()
			body, err := io.ReadAll(gz)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			r.ContentLength = int64(len(body))
			r.Body = io.NopCloser(bytes.NewBuffer(body))
		}
		next.ServeHTTP(w, r)
	})
}

func TimeTracer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tStart := time.Now()
		next.ServeHTTP(w, r)
		tEnd := time.Since(tStart)
		log.Printf("Duration for a request %s\r\n", tEnd)
	})
}

func (db *app) addCookie(w http.ResponseWriter, name, value string, key []byte) {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(value))
	signed := h.Sum(nil)
	sign := hex.EncodeToString(signed)
	cookie := http.Cookie{
		Name:   name,
		Value:  value + sign,
		MaxAge: 0,
	}
	if entry, ok := db.Data[value]; ok {
		entry.Key = string(key)
		db.Data[value] = entry
	} else {
		var entry helpers.WebData
		entry.Key = string(key)
		entry.Short = make(map[string]string)
		db.Data[value] = entry
	}
	http.SetCookie(w, &cookie)
}
func (db *app) checkCookie(cookie *http.Cookie) bool {
	data := cookie.Value[:32]
	signstring := cookie.Value[32:]
	sign, err := hex.DecodeString(signstring)
	if err != nil {
		log.Println(err)
	}
	h := hmac.New(sha256.New, []byte(db.Data[data].Key))
	h.Write([]byte(data))
	signed := h.Sum(nil)
	if hmac.Equal(sign, signed) {
		return true
	}
	return false

}

func (db *app) Cookies(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("Client_ID")
		if err != nil {
			log.Println(err)
			value := rand.RandStringRunes(32)
			key := rand.RandStringRunes(64)
			db.addCookie(w, "Client_ID", value, []byte(key))
		} else {
			if !db.checkCookie(cookie) {
				value := rand.RandStringRunes(32)
				key := rand.RandStringRunes(64)
				db.addCookie(w, "Client_ID", value, []byte(key))
			}
		}
		next.ServeHTTP(w, r)
	})
}
