package mymiddlewares

import (
	"bytes"
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// DecompressRequest - middleware для декомпрессии входящих запросов и отслеживания времени исполнения запроса
func DecompressRequestAndTimeTracer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tStart := time.Now()
		// Check for compression in request
		if (strings.Contains(r.Header.Get("Content-Encoding"), "gzip")) || (strings.Contains(r.Header.Get("Content-Encoding"), "br")) || (strings.Contains(r.Header.Get("Content-Encoding"), "deflate")) {
			defer r.Body.Close()
			// creating new zipped reader
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			defer gz.Close()
			// read unzipped body
			body, err := io.ReadAll(gz)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			// update request headers
			r.ContentLength = int64(len(body))
			r.Body = io.NopCloser(bytes.NewBuffer(body))
		}
		next.ServeHTTP(w, r)
		tEnd := time.Since(tStart)
		log.Printf("Duration for a request %s\r\n", tEnd)
	})
}
