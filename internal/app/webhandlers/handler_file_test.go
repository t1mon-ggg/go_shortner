package webhandlers

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/require"

	"github.com/t1mon-ggg/go_shortner/internal/app/config"
	"github.com/t1mon-ggg/go_shortner/internal/app/helpers"
	"github.com/t1mon-ggg/go_shortner/internal/app/rand"
	"github.com/t1mon-ggg/go_shortner/internal/app/storage"
)

func newFileServer(t *testing.T) (*cookiejar.Jar, *chi.Mux, *app) {
	jar, err := cookiejar.New(nil)
	require.NoError(t, err)
	db := NewApp()
	db.Storage = storage.NewFileDB("./createme.txt")
	db.Config = config.NewConfig()
	db.Data = make(helpers.Data)
	db.Data, err = db.Storage.Read()
	require.NoError(t, err)
	r := chi.NewRouter()
	db.MyMiddlewares(r)
	r.Route("/", db.Router)
	return jar, r, db
}

func gziped(ctype map[string]string) bool {
	for i := range ctype {
		if i == "Content-Encoding" && ctype[i] == "gzip" {
			return true
		}
	}
	return false
}

func compress(data []byte) ([]byte, error) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write([]byte(data)); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func decompress(data []byte) ([]byte, error) {
	reader := bytes.NewReader(data)
	gz, err := gzip.NewReader(reader)
	if err != nil {
		return nil, err
	}
	unzipped, err := io.ReadAll(gz)
	if err != nil {
		return nil, err
	}
	return unzipped, nil
}
func testRequest(t *testing.T, ts *httptest.Server, jar *cookiejar.Jar, method, path, body string, ctype map[string]string) (*http.Response, string) {

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Jar: jar,
	}
	var bodyreq *strings.Reader
	if gziped(ctype) {
		c, _ := compress([]byte(body))
		bodyreq = strings.NewReader(string(c))
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

// Test_defaultGetHandler - тестирование корневого хендлера
func Test_File_defaultGetHandler(t *testing.T) {
	jar, r, _ := newFileServer(t)
	ts := httptest.NewServer(r)
	defer ts.Close()
	t.Run("Test default Get handler", func(t *testing.T) {
		ctype := map[string]string{
			"Content-Type": "text/plain; charset=utf-8",
		}
		response, _ := testRequest(t, ts, jar, http.MethodGet, "/", "", ctype)
		defer response.Body.Close()
		require.Equal(t, http.StatusBadRequest, response.StatusCode)
	})
}

//Test_otherHandler - тестирование методов отличных от используемых
func Test_File_otherHandler(t *testing.T) {
	jar, r, _ := newFileServer(t)
	ts := httptest.NewServer(r)
	defer ts.Close()
	t.Run("Test other method handler", func(t *testing.T) {
		ctype := map[string]string{
			"Content-Type": "text/plain; charset=utf-8",
		}
		response, _ := testRequest(t, ts, jar, http.MethodPut, "/", "", ctype)
		defer response.Body.Close()
		require.Equal(t, http.StatusBadRequest, response.StatusCode)
	})
}

//Test_CreateShortURL - тестирование создания короткой ссылки
func Test_File_CreateShortURL(t *testing.T) {
	jar, r, _ := newFileServer(t)
	ts := httptest.NewServer(r)
	defer ts.Close()
	t.Run("Test default Get handler", func(t *testing.T) {
		ctype := map[string]string{
			"Content-Type": "text/plain; charset=utf-8",
		}
		response, body := testRequest(t, ts, jar, http.MethodPost, "/", "http://example.org/", ctype)
		defer response.Body.Close()
		require.Equal(t, http.StatusCreated, response.StatusCode)

		matched, err := regexp.Match(`\w{8}`, []byte(body))
		if err != nil {
			t.Fatal("Regexp error")
		}
		require.Equal(t, true, matched)
		err = os.Remove("./createme.txt")
		require.NoError(t, err)
	})
}

//Test_UnshortStatic - теститрование обратного преобразования короткой ссылки в исходную
func Test_File_UnshortStatic(t *testing.T) {
	type wanted struct {
		statusCode int
		body       string
	}
	type arg struct {
		method string
		query  string
		body   string
		ctype  map[string]string
		want   wanted
	}

	tests := []struct {
		name string
		args arg
	}{
		{
			name: "NotExistent",
			args: arg{
				method: http.MethodGet,
				query:  "/jdpijvHG",
				body:   "",
				ctype: map[string]string{
					"Content-Type": "text/plain; charset=utf-8",
				},

				want: wanted{
					statusCode: 404,
					body:       "",
				},
			},
		},
		{
			name: "WrongShort",
			args: arg{
				method: http.MethodGet,
				query:  "/jdpi",
				body:   "",
				ctype: map[string]string{
					"Content-Type": "text/plain; charset=utf-8",
				},

				want: wanted{
					statusCode: 400,
					body:       "",
				},
			},
		},
		{
			name: "WrongLong",
			args: arg{
				method: http.MethodGet,
				query:  "/jdpiaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				body:   "",
				ctype: map[string]string{
					"Content-Type": "text/plain; charset=utf-8",
				},

				want: wanted{
					statusCode: 400,
					body:       "",
				},
			},
		},
		{
			name: "Static",
			args: arg{
				method: http.MethodGet,
				query:  "/abcdABCD",
				body:   "",
				ctype: map[string]string{
					"Content-Type": "text/plain; charset=utf-8",
				},

				want: wanted{
					statusCode: 307,
					body:       "http://example.org",
				},
			},
		},
	}
	jar, r, db := newFileServer(t)
	db.Data["cookie1"] = helpers.WebData{Key: "secret_key", Short: map[string]string{"abcdABCD": "http://example.org"}}
	db.Storage.Write(db.Data)
	ts := httptest.NewServer(r)
	defer ts.Close()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, _ := testRequest(t, ts, jar, http.MethodGet, tt.args.query, "", tt.args.ctype)
			defer response.Body.Close()
			require.Equal(t, tt.args.want.statusCode, response.StatusCode)
			if tt.name == "Static" {
				require.Equal(t, response.Header.Get("Location"), tt.args.want.body)
			}
		})
	}
	err := os.Remove("./createme.txt")
	require.NoError(t, err)
}

//Test_File_2WayTest - теститрование обратного преобразования короткой ссылки в исходную
func Test_File_2WayTest(t *testing.T) {
	jar, r, _ := newFileServer(t)
	ts := httptest.NewServer(r)
	defer ts.Close()
	t.Run("Test 2way test", func(t *testing.T) {
		ctype := map[string]string{
			"Content-Type": "text/plain; charset=utf-8",
		}
		expected := "http://example.org"
		response1, body := testRequest(t, ts, jar, http.MethodPost, "/", expected, ctype)
		defer response1.Body.Close()
		require.Equal(t, http.StatusCreated, response1.StatusCode)
		short := strings.Split(body, "/")
		query := "/" + short[len(short)-1]
		response2, _ := testRequest(t, ts, jar, http.MethodGet, query, "", ctype)
		defer response2.Body.Close()
		require.Equal(t, http.StatusTemporaryRedirect, response2.StatusCode)
		require.Equal(t, expected, response2.Header.Get("Location"))
	})
	err := os.Remove("./createme.txt")
	require.NoError(t, err)
}

//Test_APIShort - тестирование сокращателя через API
func Test_File_APIShort(t *testing.T) {
	jar, r, _ := newFileServer(t)
	ts := httptest.NewServer(r)
	defer ts.Close()
	t.Run("Test default Get handler", func(t *testing.T) {
		ctype := map[string]string{
			"Content-Type": "application/json",
		}
		type req struct {
			URL string `json:"url"` //{"url":"<some_url>"}
		}
		s := req{URL: "http://example.org"}
		b, err := json.Marshal(s)
		require.NoError(t, err)
		response, body := testRequest(t, ts, jar, http.MethodPost, "/api/shorten", string(b), ctype)
		defer response.Body.Close()
		matched, err := regexp.Match(`{"result":\"http:\/\/\w+\.\w+\.\w\.\w:\d+\/\w{8}\"}`, []byte(body))
		require.NoError(t, err)
		require.True(t, matched)
	})
	err := os.Remove("./createme.txt")
	require.NoError(t, err)
}

//Test_API2Way - тестирование двухстороннего обмена через API
func Test_File_API2Way(t *testing.T) {
	jar, r, _ := newFileServer(t)
	ts := httptest.NewServer(r)
	defer ts.Close()
	t.Run("2Way API", func(t *testing.T) {
		ctype := map[string]string{
			"Content-Type": "application/json",
		}
		type req struct {
			URL string `json:"url"` //{"url":"<some_url>"}
		}
		type answer struct {
			A string `json:"result"` //{"result":"<short_url>"}
		}
		s := req{URL: "http://example.org"}
		b, err := json.Marshal(s)
		require.NoError(t, err)
		response, body := testRequest(t, ts, jar, http.MethodPost, "/api/shorten", string(b), ctype)
		defer response.Body.Close()
		require.Equal(t, "application/json", response.Header.Get("Content-Type"))
		require.Equal(t, http.StatusCreated, response.StatusCode)
		ss := answer{}
		err = json.Unmarshal([]byte(body), &ss)
		require.NoError(t, err)
		url := ss.A
		query := strings.Split(url, "/")
		q := fmt.Sprintf("/%s", query[len(query)-1])
		response, _ = testRequest(t, ts, jar, http.MethodGet, q, "", map[string]string{"Content-Type": "text/plain; charset=utf-8"})
		defer response.Body.Close()
		require.Equal(t, http.StatusTemporaryRedirect, response.StatusCode)
		require.Equal(t, s.URL, response.Header.Get("Location"))
	})
	err := os.Remove("./createme.txt")
	require.NoError(t, err)
}

//Test_ZippedRequest - теститрование сжатого запроса
func Test_File_ZippedRequest(t *testing.T) {
	jar, r, _ := newFileServer(t)
	ts := httptest.NewServer(r)
	defer ts.Close()
	t.Run("Test default Get handler", func(t *testing.T) {
		ctype := map[string]string{
			"Content-Type":     "application/json",
			"Content-Encoding": "gzip",
		}
		type req struct {
			URL string `json:"url"` //{"url":"<some_url>"}
		}
		s := req{URL: "http://example.org"}
		b, err := json.Marshal(s)
		require.NoError(t, err)
		response, body := testRequest(t, ts, jar, http.MethodPost, "/api/shorten", string(b), ctype)
		defer response.Body.Close()
		matched, err := regexp.Match(`{"result":\"http:\/\/\w+\.\w+\.\w\.\w:\d+\/\w{8}\"}`, []byte(body))
		require.NoError(t, err)
		require.True(t, matched)
	})
}

//Test_ZippedAnswer - тестирование сжатия ответа с сервера
func Test_File_ZippedAnswer(t *testing.T) {
	jar, r, _ := newFileServer(t)
	ts := httptest.NewServer(r)
	defer ts.Close()
	t.Run("Test default Get handler", func(t *testing.T) {
		ctype := map[string]string{
			"Content-Type":    "application/json",
			"Accept-Encoding": "gzip",
		}
		type req struct {
			URL string `json:"url"` //{"url":"<some_url>"}
		}
		s := req{URL: "http://example.org"}
		b, err := json.Marshal(s)
		require.NoError(t, err)
		response, body := testRequest(t, ts, jar, http.MethodPost, "/api/shorten", string(b), ctype)
		defer response.Body.Close()
		require.Equal(t, "gzip", response.Header.Get("Content-Encoding"))
		a, err := decompress([]byte(body))
		require.NoError(t, err)
		matched, err := regexp.Match(`{"result":\"http:\/\/\w+\.\w+\.\w\.\w:\d+\/\w{8}\"}`, a)
		require.NoError(t, err)
		require.True(t, matched)
	})
}

//Test_2WayZip - тестирование работы со сжатием данных в обе стороны
func Test_File_2WayZip(t *testing.T) {
	jar, r, _ := newFileServer(t)
	ts := httptest.NewServer(r)
	defer ts.Close()
	t.Run("Test default Get handler", func(t *testing.T) {
		ctype := map[string]string{
			"Content-Type":     "application/json",
			"Accept-Encoding":  "gzip",
			"Content-Encoding": "gzip",
		}
		type req struct {
			URL string `json:"url"` //{"url":"<some_url>"}
		}
		s := req{URL: "http://example.org"}
		b, err := json.Marshal(s)
		require.NoError(t, err)
		response, body := testRequest(t, ts, jar, http.MethodPost, "/api/shorten", string(b), ctype)
		defer response.Body.Close()
		require.Equal(t, "gzip", response.Header.Get("Content-Encoding"))
		a, err := decompress([]byte(body))
		require.NoError(t, err)
		matched, err := regexp.Match(`{"result":\"http:\/\/\w+\.\w+\.\w\.\w:\d+\/\w{8}\"}`, a)
		require.NoError(t, err)
		require.True(t, matched)
	})
}

//Test_UserURLs - тестиирование получения всех ссылок пользователя
func Test_File_UserURLs(t *testing.T) {
	type wanted struct {
		statusCode int
		body       string
	}
	type arg struct {
		method string
		query  string
		body   string
		ctype  map[string]string
		want   wanted
	}
	tests := []struct {
		name string
		args arg
	}{
		{
			name: "Not existent cookie",
			args: arg{
				method: http.MethodGet,
				query:  "/api/user/urls",
				body:   "",
				ctype: map[string]string{
					"Content-Type": "text/plain; charset=utf-8",
				},
				want: wanted{
					statusCode: http.StatusNoContent,
					body:       "",
				},
			},
		},
		{
			name: "Existent cookie",
			args: arg{
				method: http.MethodGet,
				query:  "/api/user/urls",
				body:   "",
				ctype: map[string]string{
					"Content-Type": "text/plain; charset=utf-8",
				},
				want: wanted{
					statusCode: http.StatusNoContent,
					body:       "",
				},
			},
		},
		{
			name: "Wrong cookie",
			args: arg{
				method: http.MethodGet,
				query:  "/",
				body:   "",
				ctype: map[string]string{
					"Content-Type": "text/plain; charset=utf-8",
				},
				want: wanted{
					statusCode: 0,
					body:       "",
				},
			},
		},
	}
	jar, r, _ := newFileServer(t)
	ts := httptest.NewServer(r)
	defer ts.Close()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jar, _ = cookiejar.New(nil)
			switch tt.name {
			case "Not existent cookie":
				response, _ := testRequest(t, ts, jar, http.MethodGet, tt.args.query, "", tt.args.ctype)
				defer response.Body.Close()
				require.Equal(t, http.StatusNoContent, response.StatusCode)
			case "Existent cookie":
				response, _ := testRequest(t, ts, jar, http.MethodPost, "/", "http://example.org", map[string]string{
					"Content-Type": "text/plain; charset=utf-8"})
				defer response.Body.Close()
				require.Equal(t, http.StatusCreated, response.StatusCode)
				response, _ = testRequest(t, ts, jar, http.MethodPost, "/", "http://example2.org", map[string]string{
					"Content-Type": "text/plain; charset=utf-8"})
				defer response.Body.Close()
				require.Equal(t, http.StatusCreated, response.StatusCode)
				response, body := testRequest(t, ts, jar, http.MethodGet, tt.args.query, tt.args.body, tt.args.ctype)
				defer response.Body.Close()
				require.Equal(t, "application/json", response.Header.Get("Content-Type"))
				require.Equal(t, http.StatusOK, response.StatusCode)
				type answer struct {
					Short    string `json:"short_url"`
					Original string `json:"original_url"`
				}
				d := make([]answer, 0)
				err := json.Unmarshal([]byte(body), &d)
				require.NoError(t, err)
			case "Wrong cookie":
				response, _ := testRequest(t, ts, jar, http.MethodPost, tt.args.query, "http://example3.org", tt.args.ctype)
				defer response.Body.Close()
				cookies := jar.Cookies(response.Request.URL)
				var cvalues []string
				cvalues = make([]string, 0)
				for _, c := range cookies {
					cvalues = append(cvalues, c.Value)
					c.Value = rand.RandStringRunes(96)
				}
				response, _ = testRequest(t, ts, jar, http.MethodGet, tt.args.query, "", tt.args.ctype)
				defer response.Body.Close()
				cookies = jar.Cookies(response.Request.URL)
				var mvalues []string
				mvalues = make([]string, 0)
				for _, m := range cookies {
					mvalues = append(cvalues, m.Value)
				}
				require.NotEqual(t, cvalues, mvalues)

			}
		})
	}
	err := os.Remove("./createme.txt")
	require.NoError(t, err)
}

//Test_Ping - тестирование фалового хранилища
func Test_File_Ping(t *testing.T) {
	jar, r, _ := newFileServer(t)
	ts := httptest.NewServer(r)
	defer ts.Close()
	t.Run("Test default Ping handler", func(t *testing.T) {
		ctype := map[string]string{
			"Content-Type": "text/plain; charset=utf-8",
		}
		f, err := os.Create("./createme.txt")
		require.NoError(t, err)
		f.Close()
		response, _ := testRequest(t, ts, jar, http.MethodGet, "/ping", "", ctype)
		defer response.Body.Close()
		require.Equal(t, http.StatusOK, response.StatusCode)
		err = os.Remove("./createme.txt")
		require.NoError(t, err)
		response, _ = testRequest(t, ts, jar, http.MethodGet, "/ping", "", ctype)
		defer response.Body.Close()
		require.Equal(t, http.StatusInternalServerError, response.StatusCode)
	})
}
