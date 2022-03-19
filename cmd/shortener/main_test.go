package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMyHandler(t *testing.T) {
	type want struct {
		statusCode int
	}
	type request struct {
		method string
		query  string
		body   string
	}
	tests := []struct {
		name    string
		request request
		want    want
	}{
		{
			name: "Post request",
			request: request{
				method: http.MethodPost,
				query:  "http://127.0.0.1/",
				body:   "https://yandex.ru",
			},
			want: want{
				statusCode: 201,
			},
		},
		{
			name: "Get request",
			request: request{
				method: http.MethodGet,
				query:  "http://127.0.0.1/",
				body:   "https://yandex.ru",
			},
			want: want{
				statusCode: 307,
			},
		},
		{
			name: "Other method request test",
			request: request{
				method: http.MethodPut,
				query:  "http://127.0.0.1/",
				body:   "",
			},
			want: want{
				statusCode: 400,
			},
		},
	}
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			result := func(req request) *http.Response {
				body := strings.NewReader(tt.request.body)
				request := httptest.NewRequest(tt.request.method, tt.request.query, body)
				w := httptest.NewRecorder()
				h := http.HandlerFunc(MyHandler)
				h.ServeHTTP(w, request)
				return w.Result()
			}
			switch {
			case tt.request.method == http.MethodPost:
				result := result(tt.request)
				require.Equal(t, tt.want.statusCode, result.StatusCode)
				defer result.Body.Close()
				answer, err := io.ReadAll(result.Body)
				if len(answer) == 0 {
					t.Fatal("Empty response")
				}
				if err != nil {
					t.Fatal("Response not recognized")
				}
				pattern := `http:\/\/127.0.0.1:8080\/\w{8}`
				matched, err := regexp.Match(pattern, answer)
				if err != nil {
					t.Fatal("Regexp error")
				}
				t.Log("Check response body")
				require.Equal(t, true, matched)
			case tt.request.method == http.MethodGet:
				prepare := result(tt.request)
				defer prepare.Body.Close()
				url, err := io.ReadAll(prepare.Body)
				if err != nil {
					t.Fatal(err)
				}
				request := request{
					method: http.MethodGet,
					query:  string(url),
					body:   "",
				}
				result := result(request)
				defer result.Body.Close()
				require.Equal(t, tt.want.statusCode, result.StatusCode)
				// require.Equal(t, tt.request.body, result.Header.Get("Location"))
			default:
				result := result(tt.request)
				defer result.Body.Close()
				require.Equal(t, tt.want.statusCode, result.StatusCode)
			}

		})
	}
}
