package webhandlers

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/t1mon-ggg/go_shortner/internal/app/config"
	"github.com/t1mon-ggg/go_shortner/internal/app/storage"
)

func gziped(ctype map[string]string) bool {
	for i := range ctype {
		if i == "Content-Encoding" && ctype[i] == "gzip" {
			return true
		}
	}
	return false
}

func compress(data []byte) (string, error) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write([]byte(data)); err != nil {
		return "", err
	}
	if err := gz.Close(); err != nil {
		return "", err
	}
	return string(b.Bytes()), nil
}

func testRequest(t *testing.T, ts *httptest.Server, method, path, body string, ctype map[string]string) (*http.Response, string) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	var bodyreq *strings.Reader
	if gziped(ctype) {
		c, _ := compress([]byte(body))
		bodyreq = strings.NewReader(c)
	} else {
		bodyreq = strings.NewReader(body)
	}
	req, err := http.NewRequest(method, ts.URL+path, bodyreq)
	require.NoError(t, err)
	for i := range ctype {
		req.Header.Set(i, ctype[i])
	}

	resp, err := client.Do(req)
	require.NoError(t, err)

	respBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	defer resp.Body.Close()

	return resp, string(respBody)
}

func TestDB_Router(t *testing.T) {
	db := NewApp()
	db.Storage = *storage.NewFileDB("./createme.txt")
	db.Config = *config.NewConfig()
	db.Data["ABCDabcd"] = "http://example.org"
	type want struct {
		statusCode int
		data       string
	}
	type request struct {
		method string
		query  string
		body   string
		rtype  string
		ctype  map[string]string
	}
	tests := []struct {
		name    string
		request request
		want    want
	}{
		{
			name: "Create short url test",
			request: request{
				method: http.MethodPost,
				query:  "/",
				body:   "http://example.org/",
				rtype:  "CreateShort",
				ctype: map[string]string{
					"Content-Type": "text/plain; charset=utf-8",
				},
			},
			want: want{
				statusCode: 201,
				data:       `\w{8}`,
			},
		},
		{
			name: "Unshort static url test",
			request: request{
				method: http.MethodGet,
				query:  "/ABCDabcd",
				body:   "http://example.org",
				rtype:  "GetLong",
				ctype: map[string]string{
					"Content-Type": "text/plain; charset=utf-8",
				},
			},
			want: want{
				statusCode: 307,
				data:       "http://example.org",
			},
		},
		{
			name: "2-Way test",
			request: request{
				method: http.MethodPost,
				query:  "/",
				body:   "http://example.com",
				rtype:  "2-Way",
				ctype: map[string]string{
					"Content-Type": "text/plain; charset=utf-8",
				},
			},
			want: want{
				statusCode: 307,
				data:       "http://example.com",
			},
		},
		{
			name: "Other method request test",
			request: request{
				method: http.MethodPut,
				query:  "/",
				body:   "",
				rtype:  "Other",
				ctype: map[string]string{
					"Content-Type": "text/plain; charset=utf-8",
				},
			},
			want: want{
				statusCode: 400,
				data:       "",
			},
		},
		{
			name: "Unshort not exist url",
			request: request{
				method: http.MethodGet,
				query:  "/jdpijvHG",
				body:   "",
				rtype:  "GetLong",
				ctype: map[string]string{
					"Content-Type": "text/plain; charset=utf-8",
				},
			},
			want: want{
				statusCode: 400,
				data:       "",
			},
		},
		{
			name: "Unshort wrongly short ulr ",
			request: request{
				method: http.MethodGet,
				query:  "/jvHG",
				body:   "",
				rtype:  "GetLong",
				ctype: map[string]string{
					"Content-Type": "text/plain; charset=utf-8",
				},
			},
			want: want{
				statusCode: 400,
				data:       "",
			},
		},
		{
			name: "Unshort wrongly long url",
			request: request{
				method: http.MethodGet,
				query:  "/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				body:   "",
				rtype:  "GetLong",
				ctype: map[string]string{
					"Content-Type": "",
				},
			},
			want: want{
				statusCode: 400,
				data:       "",
			},
		},
		{
			name: "Create api short url test",
			request: request{
				method: http.MethodPost,
				query:  "/api/shorten",
				body:   `{"url":"http://ip4fz0o0uwmq.yandex/zukai69rdjyqnn/ejsqdy"}`,
				rtype:  "APIShort",
				ctype: map[string]string{
					"Content-Type": "application/json",
				},
			},
			want: want{
				statusCode: 201,
				data:       `{"result":\"http:\/\/\w+\.\w+\.\w\.\w:\d+\/\w{8}\"}`,
			},
		},
		{
			name: "API 2-Way test",
			request: request{
				method: http.MethodPost,
				query:  "/api/shorten",
				body:   `{"url":"http://ip4fz0o0uwmq.yandex/zukai69rdjyqnn/ejsqdy"}`,
				rtype:  "APIShort 2-Way",
				ctype: map[string]string{
					"Content-Type": "application/json",
				},
			},
			want: want{
				statusCode: 307,
				data:       "http://ip4fz0o0uwmq.yandex/zukai69rdjyqnn/ejsqdy",
			},
		},
		{
			name: "Unshort compressed",
			request: request{
				method: http.MethodGet,
				query:  "/ABCDabcd",
				body:   "http://example.org",
				rtype:  "GetLong",
				ctype: map[string]string{
					"Content-Type":    "text/plain; charset=utf-8",
					"Accept-Encoding": "gzip, deflate, br",
				},
			},
			want: want{
				statusCode: 307,
				data:       "http://example.org",
			},
		},
		{
			name: "Create api short url test with compress",
			request: request{
				method: http.MethodPost,
				query:  "/api/shorten",
				body:   `{"url":"http://fghjt.ru"}`,
				rtype:  "APIShortCompressRequest",
				ctype: map[string]string{
					"Content-Type":     "application/json",
					"Content-Encoding": "gzip",
				},
			},
			want: want{
				statusCode: 201,
				data:       `{"result":\"http:\/\/\w+\.\w+\.\w\.\w:\d+\/\w{8}\"}`,
			},
		},
	}
	for _, tt := range tests {
		r := chi.NewRouter()
		r.Use(middleware.RequestID)
		r.Use(middleware.RealIP)
		//r.Use(middleware.Logger)
		r.Use(middleware.Recoverer)
		// r.Use(middleware.AllowContentEncoding("gzip", "br", "deflate"))
		// r.Use(middleware.Compress(5, "application/json"))
		r.Use(DecompressRequest)
		r.Route("/", db.Router)
		ts := httptest.NewServer(r)
		defer ts.Close()
		t.Run(tt.name, func(t *testing.T) {
			response, body := testRequest(t, ts, tt.request.method, tt.request.query, tt.request.body, tt.request.ctype)
			defer response.Body.Close()
			switch tt.request.rtype {
			case "CreateShort":
				require.Equal(t, tt.want.statusCode, response.StatusCode)
				matched, err := regexp.Match(tt.want.data, []byte(body))
				if err != nil {
					t.Fatal("Regexp error")
				}
				assert.Equal(t, true, matched)
			case "GetLong":
				require.Equal(t, tt.want.statusCode, response.StatusCode)
				if tt.want.statusCode != 400 {
					header := response.Header.Get("Location")
					require.Equal(t, tt.want.data, header)
				}
			case "2-Way":
				rex := regexp.MustCompile(`\w{8}`)
				short := "/" + rex.FindString(body)
				step2, _ := testRequest(t, ts, http.MethodGet, short, "", tt.request.ctype)
				defer step2.Body.Close()
				assert.Equal(t, tt.want.statusCode, step2.StatusCode)
				if tt.want.statusCode != 400 {
					header := step2.Header.Get("Location")
					require.Equal(t, tt.want.data, header)
				}
			case "APIShort":
				require.Equal(t, tt.want.statusCode, response.StatusCode)
				require.Equal(t, "application/json", response.Header.Get("Content-Type"))
				matched, err := regexp.Match(tt.want.data, []byte(body))
				if err != nil {
					t.Fatal("Regexp error")
				}
				assert.Equal(t, true, matched)
			case "APIShort 2-Way":
				type sURL struct {
					ShortURL string `json:"result"`
				}
				rex := regexp.MustCompile(`{"result":\"http:\/\/\w+\.\w+\.\w\.\w:\d+\/\w{8}\"}`)
				url := sURL{}
				err := json.Unmarshal([]byte(rex.FindString(body)), &url)
				require.NoError(t, err)
				rex = regexp.MustCompile(`\w{8}`)
				short := "/" + rex.FindString(url.ShortURL)
				step2, _ := testRequest(t, ts, http.MethodGet, short, "", map[string]string{"Content-Type": "text/plain; charset=utf-8"})
				defer step2.Body.Close()
				assert.Equal(t, tt.want.statusCode, step2.StatusCode)
				if tt.want.statusCode != 400 {
					header := step2.Header.Get("Location")
					require.Equal(t, tt.want.data, header)
				}
			case "APIShortCompress":
				require.Equal(t, tt.want.statusCode, response.StatusCode)
				require.Equal(t, "application/json", response.Header.Get("Content-Type"))
				require.NotEmpty(t, response.Header.Get("Content-Encoding"))
				rdata := strings.NewReader(body)
				r, _ := gzip.NewReader(rdata)
				s, _ := io.ReadAll(r)
				matched, err := regexp.Match(tt.want.data, []byte(s))
				if err != nil {
					t.Fatal("Regexp error")
				}
				assert.Equal(t, true, matched)
			case "APIShortCompressRequest":
				require.Equal(t, tt.want.statusCode, response.StatusCode)
				r := strings.NewReader(body)
				s, _ := io.ReadAll(r)
				fmt.Println(string(s))

			default:
				require.Equal(t, tt.want.statusCode, response.StatusCode)
			}
		})
	}
	t.Run("FileStorage Close test", func(t *testing.T) {
		db.Storage.Close()
		err := os.Remove("./createme.txt")
		require.NoError(t, err)
	})

}
