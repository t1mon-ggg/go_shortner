package webhandlers

// import (
// 	"encoding/json"
// 	"fmt"
// 	"log"
// 	"net/http"
// 	"net/http/cookiejar"
// 	"net/http/httptest"
// 	"os"
// 	"regexp"
// 	"strings"
// 	"testing"

// 	"github.com/go-chi/chi"
// 	"github.com/go-chi/chi/middleware"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"

// 	"github.com/t1mon-ggg/go_shortner/internal/app/config"
// 	"github.com/t1mon-ggg/go_shortner/internal/app/helpers"
// 	"github.com/t1mon-ggg/go_shortner/internal/app/storage"
// )

// //Test_defaultGetHandler - тестирование корневого хендлера
// func Test_defaultGetHandler(t *testing.T) {
// 	jar, err := cookiejar.New(nil)
// 	if err != nil {
// 		log.Println(err)
// 	}
// 	db := NewApp()
// 	db.Storage = storage.NewFileDB("./createme.txt")
// 	db.Config = *config.NewConfig()
// 	db.Data = make(helpers.Data)
// 	r := chi.NewRouter()
// 	r.Use(middleware.RequestID)
// 	r.Use(middleware.RealIP)
// 	r.Use(middleware.Recoverer)
// 	r.Use(middleware.AllowContentEncoding("gzip", "br", "deflate"))
// 	r.Use(middleware.Compress(5, "application/json"))
// 	r.Use(DecompressRequest)
// 	r.Use(TimeTracer)
// 	r.Use(db.Cookies)
// 	r.Route("/", db.Router)
// 	ts := httptest.NewServer(r)
// 	defer ts.Close()
// 	t.Run("Test default Get handler", func(t *testing.T) {
// 		ctype := map[string]string{
// 			"Content-Type": "text/plain; charset=utf-8",
// 		}
// 		response, _ := testRequest(t, ts, jar, http.MethodGet, "/", "", ctype)
// 		defer response.Body.Close()
// 		assert.Equal(t, http.StatusBadRequest, response.StatusCode)
// 	})
// }

// //Test_otherHandler - тестирование методов отличных от используемых
// func Test_otherHandler(t *testing.T) {
// 	jar, err := cookiejar.New(nil)
// 	if err != nil {
// 		log.Println(err)
// 	}
// 	db := NewApp()
// 	db.Storage = storage.NewFileDB("./createme.txt")
// 	db.Config = *config.NewConfig()
// 	db.Data = make(helpers.Data)
// 	r := chi.NewRouter()
// 	r.Use(middleware.RequestID)
// 	r.Use(middleware.RealIP)
// 	r.Use(middleware.Recoverer)
// 	r.Use(middleware.AllowContentEncoding("gzip", "br", "deflate"))
// 	r.Use(middleware.Compress(5, "application/json"))
// 	r.Use(DecompressRequest)
// 	r.Use(TimeTracer)
// 	r.Use(db.Cookies)
// 	r.Route("/", db.Router)
// 	ts := httptest.NewServer(r)
// 	defer ts.Close()
// 	t.Run("Test other method handler", func(t *testing.T) {
// 		ctype := map[string]string{
// 			"Content-Type": "text/plain; charset=utf-8",
// 		}
// 		response, _ := testRequest(t, ts, jar, http.MethodPut, "/", "", ctype)
// 		defer response.Body.Close()
// 		assert.Equal(t, http.StatusBadRequest, response.StatusCode)
// 	})
// }

// //Test_CreateShortURL - тестирование создания короткой ссылки
// func Test_CreateShortURL(t *testing.T) {
// 	jar, err := cookiejar.New(nil)
// 	if err != nil {
// 		log.Println(err)
// 	}
// 	db := NewApp()
// 	db.Storage = storage.NewFileDB("./createme.txt")
// 	db.Config = *config.NewConfig()
// 	db.Data = make(helpers.Data)
// 	r := chi.NewRouter()
// 	r.Use(middleware.RequestID)
// 	r.Use(middleware.RealIP)
// 	r.Use(middleware.Recoverer)
// 	r.Use(middleware.AllowContentEncoding("gzip", "br", "deflate"))
// 	r.Use(middleware.Compress(5, "application/json"))
// 	r.Use(DecompressRequest)
// 	r.Use(TimeTracer)
// 	r.Use(db.Cookies)
// 	r.Route("/", db.Router)
// 	ts := httptest.NewServer(r)
// 	defer ts.Close()
// 	t.Run("Test default Get handler", func(t *testing.T) {
// 		ctype := map[string]string{
// 			"Content-Type": "text/plain; charset=utf-8",
// 		}
// 		response, body := testRequest(t, ts, jar, http.MethodPost, "/", "http://example.org/", ctype)
// 		defer response.Body.Close()
// 		require.Equal(t, http.StatusCreated, response.StatusCode)

// 		matched, err := regexp.Match(`\w{8}`, []byte(body))
// 		if err != nil {
// 			t.Fatal("Regexp error")
// 		}
// 		require.Equal(t, true, matched)
// 		err = os.Remove("./createme.txt")
// 		require.NoError(t, err)
// 	})
// }

// //Test_UnshortStatic - теститрование обратного преобразования короткой ссылки в исходную
// func Test_UnshortStatic(t *testing.T) {
// 	type wanted struct {
// 		statusCode int
// 		body       string
// 	}
// 	type arg struct {
// 		method string
// 		query  string
// 		body   string
// 		ctype  map[string]string
// 		want   wanted
// 	}

// 	tests := []struct {
// 		name string
// 		args arg
// 	}{
// 		{
// 			name: "NotExistent",
// 			args: arg{
// 				method: http.MethodGet,
// 				query:  "/jdpijvHG",
// 				body:   "",
// 				ctype: map[string]string{
// 					"Content-Type": "text/plain; charset=utf-8",
// 				},

// 				want: wanted{
// 					statusCode: 404,
// 					body:       "",
// 				},
// 			},
// 		},
// 		{
// 			name: "WrongShort",
// 			args: arg{
// 				method: http.MethodGet,
// 				query:  "/jdpi",
// 				body:   "",
// 				ctype: map[string]string{
// 					"Content-Type": "text/plain; charset=utf-8",
// 				},

// 				want: wanted{
// 					statusCode: 400,
// 					body:       "",
// 				},
// 			},
// 		},
// 		{
// 			name: "WrongLong",
// 			args: arg{
// 				method: http.MethodGet,
// 				query:  "/jdpiaaaaaaaaaaaaaaaaaaaaaaaaaaa",
// 				body:   "",
// 				ctype: map[string]string{
// 					"Content-Type": "text/plain; charset=utf-8",
// 				},

// 				want: wanted{
// 					statusCode: 400,
// 					body:       "",
// 				},
// 			},
// 		},
// 		{
// 			name: "Static",
// 			args: arg{
// 				method: http.MethodGet,
// 				query:  "/abcdABCD",
// 				body:   "",
// 				ctype: map[string]string{
// 					"Content-Type": "text/plain; charset=utf-8",
// 				},

// 				want: wanted{
// 					statusCode: 307,
// 					body:       "http://example.org",
// 				},
// 			},
// 		},
// 	}

// 	jar, err := cookiejar.New(nil)
// 	if err != nil {
// 		log.Println(err)
// 	}
// 	db := NewApp()
// 	db.Storage = storage.NewFileDB("./createme.txt")
// 	db.Config = *config.NewConfig()
// 	db.Data = make(helpers.Data)
// 	db.Data["cookie1"] = helpers.WebData{Key: "secret_key", Short: map[string]string{"abcdABCD": "http://example.org"}}
// 	db.Storage.Write(db.Data)
// 	r := chi.NewRouter()
// 	r.Use(middleware.RequestID)
// 	r.Use(middleware.RealIP)
// 	r.Use(middleware.Recoverer)
// 	r.Use(middleware.AllowContentEncoding("gzip", "br", "deflate"))
// 	r.Use(middleware.Compress(5, "application/json"))
// 	r.Use(DecompressRequest)
// 	r.Use(TimeTracer)
// 	r.Use(db.Cookies)
// 	r.Route("/", db.Router)
// 	ts := httptest.NewServer(r)
// 	defer ts.Close()
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			response, _ := testRequest(t, ts, jar, http.MethodGet, tt.args.query, "", tt.args.ctype)
// 			defer response.Body.Close()
// 			require.Equal(t, tt.args.want.statusCode, response.StatusCode)
// 			if tt.name == "Static" {
// 				require.Equal(t, response.Header.Get("Location"), tt.args.want.body)
// 			}
// 		})
// 	}
// 	err = os.Remove("./createme.txt")
// 	require.NoError(t, err)
// }

// //Test_UnshortStatic - теститрование обратного преобразования короткой ссылки в исходную
// func Test_2WayTest(t *testing.T) {
// 	jar, err := cookiejar.New(nil)
// 	if err != nil {
// 		log.Println(err)
// 	}
// 	db := NewApp()
// 	db.Storage = storage.NewFileDB("./createme.txt")
// 	db.Config = *config.NewConfig()
// 	db.Data = make(helpers.Data)
// 	db.Storage.Write(db.Data)
// 	r := chi.NewRouter()
// 	r.Use(middleware.RequestID)
// 	r.Use(middleware.RealIP)
// 	r.Use(middleware.Recoverer)
// 	r.Use(middleware.AllowContentEncoding("gzip", "br", "deflate"))
// 	r.Use(middleware.Compress(5, "application/json"))
// 	r.Use(DecompressRequest)
// 	r.Use(TimeTracer)
// 	r.Use(db.Cookies)
// 	r.Route("/", db.Router)
// 	ts := httptest.NewServer(r)
// 	defer ts.Close()
// 	t.Run("Test 2way test", func(t *testing.T) {
// 		ctype := map[string]string{
// 			"Content-Type": "text/plain; charset=utf-8",
// 		}
// 		expected := "http://example.org"
// 		response1, body := testRequest(t, ts, jar, http.MethodPost, "/", expected, ctype)
// 		defer response1.Body.Close()
// 		require.Equal(t, http.StatusCreated, response1.StatusCode)
// 		short := strings.Split(body, "/")
// 		query := "/" + short[len(short)-1]
// 		response2, _ := testRequest(t, ts, jar, http.MethodGet, query, "", ctype)
// 		defer response2.Body.Close()
// 		require.Equal(t, http.StatusTemporaryRedirect, response2.StatusCode)
// 		require.Equal(t, expected, response2.Header.Get("Location"))
// 		err := os.Remove("./createme.txt")
// 		require.NoError(t, err)
// 	})
// }

// //Test_APIShort - тестирование сокращателя через API
// func Test_APIShort(t *testing.T) {
// 	jar, err := cookiejar.New(nil)
// 	if err != nil {
// 		log.Println(err)
// 	}
// 	db := NewApp()
// 	db.Storage = storage.NewFileDB("./createme.txt")
// 	db.Config = *config.NewConfig()
// 	db.Data = make(helpers.Data)
// 	r := chi.NewRouter()
// 	r.Use(middleware.RequestID)
// 	r.Use(middleware.RealIP)
// 	r.Use(middleware.Recoverer)
// 	r.Use(middleware.AllowContentEncoding("gzip", "br", "deflate"))
// 	r.Use(middleware.Compress(5, "application/json"))
// 	r.Use(DecompressRequest)
// 	r.Use(TimeTracer)
// 	r.Use(db.Cookies)
// 	r.Route("/", db.Router)
// 	ts := httptest.NewServer(r)
// 	defer ts.Close()
// 	t.Run("Test default Get handler", func(t *testing.T) {
// 		ctype := map[string]string{
// 			"Content-Type": "application/json",
// 		}
// 		type req struct {
// 			URL string `json:"url"` //{"url":"<some_url>"}
// 		}
// 		s := req{URL: "http://example.org"}
// 		b, err := json.Marshal(s)
// 		require.NoError(t, err)
// 		response, body := testRequest(t, ts, jar, http.MethodPost, "/api/shorten", string(b), ctype)
// 		defer response.Body.Close()
// 		matched, err := regexp.Match(`{"result":\"http:\/\/\w+\.\w+\.\w\.\w:\d+\/\w{8}\"}`, []byte(body))
// 		require.NoError(t, err)
// 		require.True(t, matched)
// 	})
// 	err = os.Remove("./createme.txt")
// 	require.NoError(t, err)
// }

// //Test_API2Way - тестирование двухстороннего обмена через API
// func Test_API2Way(t *testing.T) {
// 	jar, err := cookiejar.New(nil)
// 	if err != nil {
// 		log.Println(err)
// 	}
// 	db := NewApp()
// 	db.Storage = storage.NewFileDB("./createme.txt")
// 	db.Config = *config.NewConfig()
// 	db.Data = make(helpers.Data)
// 	r := chi.NewRouter()
// 	r.Use(middleware.RequestID)
// 	r.Use(middleware.RealIP)
// 	r.Use(middleware.Recoverer)
// 	r.Use(middleware.AllowContentEncoding("gzip", "br", "deflate"))
// 	r.Use(middleware.Compress(5, "application/json"))
// 	r.Use(DecompressRequest)
// 	r.Use(TimeTracer)
// 	r.Use(db.Cookies)
// 	r.Route("/", db.Router)
// 	ts := httptest.NewServer(r)
// 	defer ts.Close()
// 	t.Run("2Way API", func(t *testing.T) {
// 		ctype := map[string]string{
// 			"Content-Type": "application/json",
// 		}
// 		type req struct {
// 			URL string `json:"url"` //{"url":"<some_url>"}
// 		}
// 		type answer struct {
// 			A string `json:"result"` //{"result":"<short_url>"}
// 		}
// 		s := req{URL: "http://example.org"}
// 		b, err := json.Marshal(s)
// 		require.NoError(t, err)
// 		response, body := testRequest(t, ts, jar, http.MethodPost, "/api/shorten", string(b), ctype)
// 		defer response.Body.Close()
// 		require.Equal(t, "application/json", response.Header.Get("Content-Type"))
// 		require.Equal(t, http.StatusCreated, response.StatusCode)
// 		ss := answer{}
// 		err = json.Unmarshal([]byte(body), &ss)
// 		require.NoError(t, err)
// 		url := ss.A
// 		query := strings.Split(url, "/")
// 		q := fmt.Sprintf("/%s", query[len(query)-1])
// 		response, _ = testRequest(t, ts, jar, http.MethodGet, q, "", map[string]string{"Content-Type": "text/plain; charset=utf-8"})
// 		defer response.Body.Close()
// 		require.Equal(t, http.StatusTemporaryRedirect, response.StatusCode)
// 		require.Equal(t, s.URL, response.Header.Get("Location"))
// 	})
// 	err = os.Remove("./createme.txt")
// 	require.NoError(t, err)
// }

// //Test_ZippedRequest - теститрование сжатого запроса
// func Test_ZippedRequest(t *testing.T) {
// 	jar, err := cookiejar.New(nil)
// 	if err != nil {
// 		log.Println(err)
// 	}
// 	db := NewApp()
// 	db.Storage = storage.NewFileDB("./createme.txt")
// 	db.Config = *config.NewConfig()
// 	db.Data = make(helpers.Data)
// 	r := chi.NewRouter()
// 	r.Use(middleware.RequestID)
// 	r.Use(middleware.RealIP)
// 	r.Use(middleware.Recoverer)
// 	r.Use(middleware.AllowContentEncoding("gzip", "br", "deflate"))
// 	r.Use(middleware.Compress(5, "application/json"))
// 	r.Use(DecompressRequest)
// 	r.Use(TimeTracer)
// 	r.Use(db.Cookies)
// 	r.Route("/", db.Router)
// 	ts := httptest.NewServer(r)
// 	defer ts.Close()
// 	t.Run("Test default Get handler", func(t *testing.T) {
// 		ctype := map[string]string{
// 			"Content-Type":     "application/json",
// 			"Content-Encoding": "gzip",
// 		}
// 		type req struct {
// 			URL string `json:"url"` //{"url":"<some_url>"}
// 		}
// 		s := req{URL: "http://example.org"}
// 		b, err := json.Marshal(s)
// 		require.NoError(t, err)
// 		response, body := testRequest(t, ts, jar, http.MethodPost, "/api/shorten", string(b), ctype)
// 		defer response.Body.Close()
// 		matched, err := regexp.Match(`{"result":\"http:\/\/\w+\.\w+\.\w\.\w:\d+\/\w{8}\"}`, []byte(body))
// 		require.NoError(t, err)
// 		require.True(t, matched)
// 	})
// }

// //Test_ZippedAnswer - тестирование сжатия ответа с сервера
// func Test_ZippedAnswer(t *testing.T) {
// 	jar, err := cookiejar.New(nil)
// 	if err != nil {
// 		log.Println(err)
// 	}
// 	db := NewApp()
// 	db.Storage = storage.NewFileDB("./createme.txt")
// 	db.Config = *config.NewConfig()
// 	db.Data = make(helpers.Data)
// 	r := chi.NewRouter()
// 	r.Use(middleware.RequestID)
// 	r.Use(middleware.RealIP)
// 	r.Use(middleware.Recoverer)
// 	r.Use(middleware.AllowContentEncoding("gzip", "br", "deflate"))
// 	r.Use(middleware.Compress(5, "application/json"))
// 	r.Use(DecompressRequest)
// 	r.Use(TimeTracer)
// 	r.Use(db.Cookies)
// 	r.Route("/", db.Router)
// 	ts := httptest.NewServer(r)
// 	defer ts.Close()
// 	t.Run("Test default Get handler", func(t *testing.T) {
// 		ctype := map[string]string{
// 			"Content-Type":    "application/json",
// 			"Accept-Encoding": "gzip",
// 		}
// 		type req struct {
// 			URL string `json:"url"` //{"url":"<some_url>"}
// 		}
// 		s := req{URL: "http://example.org"}
// 		b, err := json.Marshal(s)
// 		require.NoError(t, err)
// 		response, body := testRequest(t, ts, jar, http.MethodPost, "/api/shorten", string(b), ctype)
// 		defer response.Body.Close()
// 		require.Equal(t, "gzip", response.Header.Get("Content-Encoding"))
// 		a, err := decompress([]byte(body))
// 		require.NoError(t, err)
// 		matched, err := regexp.Match(`{"result":\"http:\/\/\w+\.\w+\.\w\.\w:\d+\/\w{8}\"}`, a)
// 		require.NoError(t, err)
// 		require.True(t, matched)
// 	})
// }

// //Test_2WayZip - тестирование работы сос жатием данных в обе стороны
// func Test_2WayZip(t *testing.T) {
// 	jar, err := cookiejar.New(nil)
// 	if err != nil {
// 		log.Println(err)
// 	}
// 	db := NewApp()
// 	db.Storage = storage.NewFileDB("./createme.txt")
// 	db.Config = *config.NewConfig()
// 	db.Data = make(helpers.Data)
// 	r := chi.NewRouter()
// 	r.Use(middleware.RequestID)
// 	r.Use(middleware.RealIP)
// 	r.Use(middleware.Recoverer)
// 	r.Use(middleware.AllowContentEncoding("gzip", "br", "deflate"))
// 	r.Use(middleware.Compress(5, "application/json"))
// 	r.Use(DecompressRequest)
// 	r.Use(TimeTracer)
// 	r.Use(db.Cookies)
// 	r.Route("/", db.Router)
// 	ts := httptest.NewServer(r)
// 	defer ts.Close()
// 	t.Run("Test default Get handler", func(t *testing.T) {
// 		ctype := map[string]string{
// 			"Content-Type":     "application/json",
// 			"Accept-Encoding":  "gzip",
// 			"Content-Encoding": "gzip",
// 		}
// 		type req struct {
// 			URL string `json:"url"` //{"url":"<some_url>"}
// 		}
// 		s := req{URL: "http://example.org"}
// 		b, err := json.Marshal(s)
// 		require.NoError(t, err)
// 		response, body := testRequest(t, ts, jar, http.MethodPost, "/api/shorten", string(b), ctype)
// 		defer response.Body.Close()
// 		require.Equal(t, "gzip", response.Header.Get("Content-Encoding"))
// 		a, err := decompress([]byte(body))
// 		require.NoError(t, err)
// 		matched, err := regexp.Match(`{"result":\"http:\/\/\w+\.\w+\.\w\.\w:\d+\/\w{8}\"}`, a)
// 		require.NoError(t, err)
// 		require.True(t, matched)
// 	})
// }

// //Test_UserURLs - тестиирование получения всех ссылок пользователя
// func Test_UserURLs(t *testing.T) {
// 	type wanted struct {
// 		statusCode int
// 		body       string
// 	}
// 	type arg struct {
// 		method string
// 		query  string
// 		body   string
// 		ctype  map[string]string
// 		want   wanted
// 	}
// 	tests := []struct {
// 		name string
// 		args arg
// 	}{
// 		{
// 			name: "Not existent cookie",
// 			args: arg{
// 				method: http.MethodGet,
// 				query:  "/api/user/urls",
// 				body:   "",
// 				ctype: map[string]string{
// 					"Content-Type": "text/plain; charset=utf-8",
// 				},
// 				want: wanted{
// 					statusCode: http.StatusNoContent,
// 					body:       "",
// 				},
// 			},
// 		},
// 		{
// 			name: "Existent cookie",
// 			args: arg{
// 				method: http.MethodGet,
// 				query:  "/api/user/urls",
// 				body:   "",
// 				ctype: map[string]string{
// 					"Content-Type": "text/plain; charset=utf-8",
// 				},
// 				want: wanted{
// 					statusCode: http.StatusNoContent,
// 					body:       "",
// 				},
// 			},
// 		},
// 	}
// 	jar1, err := cookiejar.New(nil)
// 	if err != nil {
// 		log.Println(err)
// 	}
// 	jar2, err := cookiejar.New(nil)
// 	if err != nil {
// 		log.Println(err)
// 	}
// 	db := NewApp()
// 	db.Storage = storage.NewFileDB("./createme.txt")
// 	db.Config = *config.NewConfig()
// 	db.Data = make(helpers.Data)
// 	r := chi.NewRouter()
// 	r.Use(middleware.RequestID)
// 	r.Use(middleware.RealIP)
// 	r.Use(middleware.Recoverer)
// 	r.Use(middleware.AllowContentEncoding("gzip", "br", "deflate"))
// 	r.Use(middleware.Compress(5, "application/json"))
// 	r.Use(DecompressRequest)
// 	r.Use(TimeTracer)
// 	r.Use(db.Cookies)
// 	r.Route("/", db.Router)
// 	ts := httptest.NewServer(r)
// 	defer ts.Close()
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			switch tt.name {
// 			case "Not existent cookie":
// 				response, _ := testRequest(t, ts, jar1, http.MethodGet, "/api/user/urls", "", tt.args.ctype)
// 				defer response.Body.Close()
// 				assert.Equal(t, http.StatusNoContent, response.StatusCode)
// 			case "Existent cookie":
// 				response, _ := testRequest(t, ts, jar2, http.MethodPost, "/", "http://example.org", map[string]string{
// 					"Content-Type": "text/plain; charset=utf-8"})
// 				defer response.Body.Close()
// 				assert.Equal(t, http.StatusCreated, response.StatusCode)
// 				response, _ = testRequest(t, ts, jar2, http.MethodPost, "/", "http://example2.org", map[string]string{
// 					"Content-Type": "text/plain; charset=utf-8"})
// 				defer response.Body.Close()
// 				assert.Equal(t, http.StatusCreated, response.StatusCode)
// 				response, body := testRequest(t, ts, jar2, http.MethodGet, "/api/user/urls", tt.args.body, tt.args.ctype)
// 				defer response.Body.Close()
// 				assert.Equal(t, "application/json", response.Header.Get("Content-Type"))
// 				assert.Equal(t, http.StatusOK, response.StatusCode)
// 				type answer struct {
// 					Short    string `json:"short_url"`
// 					Original string `json:"original_url"`
// 				}
// 				d := make([]answer, 0)
// 				err = json.Unmarshal([]byte(body), &d)
// 				require.NoError(t, err)
// 			}
// 		})
// 	}
// 	err = os.Remove("./createme.txt")
// 	require.NoError(t, err)
// }

// //Test_Ping - тестиирование получения всех ссылок пользователя
// func Test_Ping(t *testing.T) {
// 	jar, err := cookiejar.New(nil)
// 	if err != nil {
// 		log.Println(err)
// 	}
// 	db := NewApp()
// 	db.Storage = storage.NewFileDB("./createme.txt")
// 	db.Config = *config.NewConfig()
// 	db.Data = make(helpers.Data)
// 	r := chi.NewRouter()
// 	r.Use(middleware.RequestID)
// 	r.Use(middleware.RealIP)
// 	r.Use(middleware.Recoverer)
// 	r.Use(middleware.AllowContentEncoding("gzip", "br", "deflate"))
// 	r.Use(middleware.Compress(5, "application/json"))
// 	r.Use(DecompressRequest)
// 	r.Use(TimeTracer)
// 	r.Use(db.Cookies)
// 	r.Route("/", db.Router)
// 	ts := httptest.NewServer(r)
// 	defer ts.Close()
// 	t.Run("Test default Ping handler", func(t *testing.T) {
// 		ctype := map[string]string{
// 			"Content-Type": "text/plain; charset=utf-8",
// 		}
// 		response, _ := testRequest(t, ts, jar, http.MethodGet, "/ping", "", ctype)
// 		defer response.Body.Close()
// 		assert.Equal(t, http.StatusNoContent, response.StatusCode)
// 	})
// 	err = os.Remove("./createme.txt")
// 	require.NoError(t, err)
// }
