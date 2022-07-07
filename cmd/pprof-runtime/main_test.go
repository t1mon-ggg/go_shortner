package main_test

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/t1mon-ggg/go_shortner/app/helpers"
)

var cmd *exec.Cmd

const host = `127.0.0.1:9090`

func init() {
	build := exec.Command("go", "build", "-o", "./pprof-runtime-test")
	err := build.Run()
	if err != nil {
		log.Fatal("build: ", err)
	}
	cmd = exec.Command("./pprof-runtime-test", "-a", host, "-b", "http://"+host, "-d", "postgresql://postgres:postgrespw@127.0.0.1:5432/praktikum?sslmode=disable", "-memprofile", "profiles/mem.pprof", "-cpuprofile", "profiles/cpu.pprof", ">", "exec.log", "2>&1")
	go func(*exec.Cmd) {
		log.Println("Args:", cmd.Args)
		log.Println("Path:", cmd.Path)
		err = cmd.Start()
		if err != nil {
			log.Fatal("Run ", err)
		}
	}(cmd)
	time.Sleep(5 * time.Second)
}

func newJar(t *testing.T) *cookiejar.Jar {
	jar, err := cookiejar.New(nil)
	require.NoError(t, err)
	return jar
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

func testRequest(t *testing.T, jar *cookiejar.Jar, method, path, body string, ctype map[string]string) (*http.Response, string) {
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
	req, err := http.NewRequest(method, fmt.Sprintf("http://%s", host)+path, bodyreq)
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
func Test_defaultGetHandler(t *testing.T) {
	jar := newJar(t)
	t.Run("Test default Get handler", func(t *testing.T) {
		ctype := map[string]string{
			"Content-Type": "text/plain; charset=utf-8",
		}
		response, _ := testRequest(t, jar, http.MethodGet, "/", "", ctype)
		defer response.Body.Close()
		require.Equal(t, http.StatusBadRequest, response.StatusCode)
	})
}

//Test_otherHandler - тестирование методов отличных от используемых
func Test_otherHandler(t *testing.T) {
	jar := newJar(t)
	t.Run("Test other method handler", func(t *testing.T) {
		ctype := map[string]string{
			"Content-Type": "text/plain; charset=utf-8",
		}
		response, _ := testRequest(t, jar, http.MethodPut, "/", "", ctype)
		defer response.Body.Close()
		require.Equal(t, http.StatusBadRequest, response.StatusCode)
	})
}

//Test_CreateShortURL - тестирование создания короткой ссылки
func Test_CreateShortURL(t *testing.T) {
	jar := newJar(t)
	t.Run("Test default Get handler", func(t *testing.T) {
		ctype := map[string]string{
			"Content-Type": "text/plain; charset=utf-8",
		}
		response, body := testRequest(t, jar, http.MethodPost, "/", "http://example.org/", ctype)
		defer response.Body.Close()
		require.Equal(t, http.StatusCreated, response.StatusCode)

		matched, err := regexp.Match(`\w{8}`, []byte(body))
		if err != nil {
			t.Fatal("Regexp error")
		}
		require.Equal(t, true, matched)
	})
}

//Test_UnshortStatic - теститрование обратного преобразования короткой ссылки в исходную
func Test_UnshortStatic(t *testing.T) {
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
	jar := newJar(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, _ := testRequest(t, jar, http.MethodGet, tt.args.query, "", tt.args.ctype)
			defer response.Body.Close()
			require.Equal(t, tt.args.want.statusCode, response.StatusCode)
			if tt.name == "Static" {
				require.Equal(t, response.Header.Get("Location"), tt.args.want.body)
			}
		})
	}
}

//Test_2WayTest - теститрование обратного преобразования короткой ссылки в исходную
func Test_2WayTest(t *testing.T) {
	jar := newJar(t)
	t.Run("Test 2way test", func(t *testing.T) {
		ctype := map[string]string{
			"Content-Type": "text/plain; charset=utf-8",
		}
		expected := "http://example.org"
		response1, body := testRequest(t, jar, http.MethodPost, "/", expected, ctype)
		defer response1.Body.Close()
		require.Equal(t, http.StatusCreated, response1.StatusCode)
		short := strings.Split(body, "/")
		query := "/" + short[len(short)-1]
		response2, _ := testRequest(t, jar, http.MethodGet, query, "", ctype)
		defer response2.Body.Close()
		require.Equal(t, http.StatusTemporaryRedirect, response2.StatusCode)
		require.Equal(t, expected, response2.Header.Get("Location"))
	})
}

//Test_APIShort - тестирование сокращателя через API
func Test_APIShort(t *testing.T) {
	jar := newJar(t)
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
		response, body := testRequest(t, jar, http.MethodPost, "/api/shorten", string(b), ctype)
		defer response.Body.Close()
		matched, err := regexp.Match(`{"result":\"http:\/\/\w+\.\w+\.\w\.\w:\d+\/\w{8}\"}`, []byte(body))
		require.NoError(t, err)
		require.True(t, matched)
	})
}

//Test_API2Way - тестирование двухстороннего обмена через API
func Test_API2Way(t *testing.T) {
	jar := newJar(t)
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
		response, body := testRequest(t, jar, http.MethodPost, "/api/shorten", string(b), ctype)
		defer response.Body.Close()
		require.Equal(t, "application/json", response.Header.Get("Content-Type"))
		require.Equal(t, http.StatusCreated, response.StatusCode)
		ss := answer{}
		err = json.Unmarshal([]byte(body), &ss)
		require.NoError(t, err)
		url := ss.A
		query := strings.Split(url, "/")
		q := fmt.Sprintf("/%s", query[len(query)-1])
		response, _ = testRequest(t, jar, http.MethodGet, q, "", map[string]string{"Content-Type": "text/plain; charset=utf-8"})
		defer response.Body.Close()
		require.Equal(t, http.StatusTemporaryRedirect, response.StatusCode)
		require.Equal(t, s.URL, response.Header.Get("Location"))
	})
}

//Test_ZippedRequest - теститрование сжатого запроса
func Test_ZippedRequest(t *testing.T) {
	jar := newJar(t)
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
		response, body := testRequest(t, jar, http.MethodPost, "/api/shorten", string(b), ctype)
		defer response.Body.Close()
		matched, err := regexp.Match(`{"result":\"http:\/\/\w+\.\w+\.\w\.\w:\d+\/\w{8}\"}`, []byte(body))
		require.NoError(t, err)
		require.True(t, matched)
	})
}

//Test_ZippedAnswer - тестирование сжатия ответа с сервера
func Test_ZippedAnswer(t *testing.T) {
	jar := newJar(t)
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
		response, body := testRequest(t, jar, http.MethodPost, "/api/shorten", string(b), ctype)
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
func Test_2WayZip(t *testing.T) {
	jar := newJar(t)
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
		response, body := testRequest(t, jar, http.MethodPost, "/api/shorten", string(b), ctype)
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
func Test_UserURLs(t *testing.T) {
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
	jar := newJar(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jar, _ = cookiejar.New(nil)
			switch tt.name {
			case "Not existent cookie":
				response, _ := testRequest(t, jar, http.MethodGet, tt.args.query, "", tt.args.ctype)
				defer response.Body.Close()
				require.Equal(t, http.StatusNoContent, response.StatusCode)
			case "Existent cookie":
				response, _ := testRequest(t, jar, http.MethodPost, "/", "http://example.org", map[string]string{
					"Content-Type": "text/plain; charset=utf-8"})
				defer response.Body.Close()
				require.Equal(t, http.StatusCreated, response.StatusCode)
				response, _ = testRequest(t, jar, http.MethodPost, "/", "http://example2.org", map[string]string{
					"Content-Type": "text/plain; charset=utf-8"})
				defer response.Body.Close()
				require.Equal(t, http.StatusCreated, response.StatusCode)
				response, body := testRequest(t, jar, http.MethodGet, tt.args.query, tt.args.body, tt.args.ctype)
				defer response.Body.Close()
				require.Equal(t, http.StatusOK, response.StatusCode)
				require.Equal(t, "application/json", response.Header.Get("Content-Type"))
				type answer struct {
					Short    string `json:"short_url"`
					Original string `json:"original_url"`
				}
				d := make([]answer, 0)
				err := json.Unmarshal([]byte(body), &d)
				require.NoError(t, err)
			case "Wrong cookie":
				response, _ := testRequest(t, jar, http.MethodPost, tt.args.query, "http://example3.org", tt.args.ctype)
				defer response.Body.Close()
				cookies := jar.Cookies(response.Request.URL)
				var cvalues []string
				cvalues = make([]string, 0)
				for _, c := range cookies {
					cvalues = append(cvalues, c.Value)
					c.Value = helpers.RandStringRunes(96)
				}
				response, _ = testRequest(t, jar, http.MethodGet, tt.args.query, "", tt.args.ctype)
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
}

//Test_Ping - тестирование хранилища
func Test_Ping(t *testing.T) {
	jar := newJar(t)
	t.Run("Test Ping handler", func(t *testing.T) {
		ctype := map[string]string{
			"Content-Type": "text/plain; charset=utf-8",
		}
		response, _ := testRequest(t, jar, http.MethodGet, "/ping", "", ctype)
		defer response.Body.Close()
		require.Equal(t, http.StatusOK, response.StatusCode)
	})
}

//Test_Delete - удаления тегов
func Test_Delete(t *testing.T) {
	jar := newJar(t)
	type input struct {
		Correlation string `json:"correlation_id"`
		Long        string `json:"original_url"`
	}
	type output struct {
		Correlation string `json:"correlation_id"`
		Short       string `json:"short_url"`
	}
	response, _ := testRequest(t, jar, http.MethodPost, "/", "", map[string]string{"Content-Type": "text/plain; charset=utf-8"})
	defer response.Body.Close()
	in := []input{
		{Correlation: "12345",
			Long: "http://example1.org"},
		{Correlation: "12345",
			Long: "http://example2.org"},
		{Correlation: "12345",
			Long: "http://example3.org"},
		{Correlation: "12345",
			Long: "http://example4.org"},
		{Correlation: "12345",
			Long: "http://example5.org"},
		{Correlation: "12345",
			Long: "http://example6.org"},
		{Correlation: "12345",
			Long: "http://example7.org"},
		{Correlation: "12345",
			Long: "http://example8.org"},
		{Correlation: "12345",
			Long: "http://example9.org"},
		{Correlation: "12345",
			Long: "http://example10.org"},
		{Correlation: "12345",
			Long: "http://example11.org"},
		{Correlation: "12345",
			Long: "http://example12.org"},
		{Correlation: "12345",
			Long: "http://example13.org"},
		{Correlation: "12345",
			Long: "http://example14.org"},
		{Correlation: "12345",
			Long: "http://example15.org"},
		{Correlation: "12345",
			Long: "http://example16.org"},
		{Correlation: "12345",
			Long: "http://example17.org"},
		{Correlation: "12345",
			Long: "http://example18.org"},
		{Correlation: "12345",
			Long: "http://example19.org"},
		{Correlation: "12345",
			Long: "http://example20.org"},
	}
	ctype := map[string]string{
		"Content-Type": "application/json",
	}
	body, err := json.Marshal(in)
	require.NoError(t, err)
	response, astring := testRequest(t, jar, http.MethodPost, "/api/shorten/batch", string(body), ctype)
	defer response.Body.Close()
	require.Equal(t, http.StatusCreated, response.StatusCode)
	var answer []output
	require.NoError(t, err)
	err = json.Unmarshal([]byte(astring), &answer)
	require.NoError(t, err)
	require.Equal(t, len(in), len(answer))
	del := make([]string, 0)
	for i := 5; i < 12; i++ {
		re := regexp.MustCompile(`\w{8}`)
		match := re.FindString(answer[i].Short)
		del = append(del, match)
	}
	for i := range del {
		del[i] = fmt.Sprintf("\"%s\"", del[i])
	}
	s := strings.Join(del, ",")
	s = "[" + s + "]"
	response, _ = testRequest(t, jar, http.MethodDelete, "/api/user/urls", s, ctype)
	defer response.Body.Close()
	require.Equal(t, http.StatusAccepted, response.StatusCode)
	time.Sleep(25 * time.Second)
	ctype = map[string]string{
		"Content-Type": "text/plain; charset=utf-8",
	}
	for i := 5; i < 12; i++ {
		re := regexp.MustCompile(`\w{8}`)
		match := re.FindString(answer[i].Short)
		response, _ = testRequest(t, jar, http.MethodGet, fmt.Sprintf("/%s", match), s, ctype)
		defer response.Body.Close()
		require.Equal(t, http.StatusGone, response.StatusCode)
	}
}

//Test_BatchAPI - массовое заполнение базы
func Test_BatchAPI(t *testing.T) {
	jar := newJar(t)
	type input struct {
		Correlation string `json:"correlation_id"`
		Long        string `json:"original_url"`
	}
	type output struct {
		Correlation string `json:"correlation_id"`
		Short       string `json:"short_url"`
	}

	in := []input{
		{Correlation: "12345",
			Long: "http://example1.org"},
		{Correlation: "12345",
			Long: "http://example2.org"},
		{Correlation: "12345",
			Long: "http://example3.org"},
		{Correlation: "12345",
			Long: "http://example4.org"},
		{Correlation: "12345",
			Long: "http://example5.org"},
		{Correlation: "12345",
			Long: "http://example6.org"},
		{Correlation: "12345",
			Long: "http://example7.org"},
		{Correlation: "12345",
			Long: "http://example8.org"},
		{Correlation: "12345",
			Long: "http://example9.org"},
		{Correlation: "12345",
			Long: "http://example10.org"},
		{Correlation: "12345",
			Long: "http://example11.org"},
		{Correlation: "12345",
			Long: "http://example12.org"},
		{Correlation: "12345",
			Long: "http://example13.org"},
		{Correlation: "12345",
			Long: "http://example14.org"},
		{Correlation: "12345",
			Long: "http://example15.org"},
		{Correlation: "12345",
			Long: "http://example16.org"},
		{Correlation: "12345",
			Long: "http://example17.org"},
		{Correlation: "12345",
			Long: "http://example18.org"},
		{Correlation: "12345",
			Long: "http://example19.org"},
		{Correlation: "12345",
			Long: "http://example20.org"},
	}
	ctype := map[string]string{
		"Content-Type": "application/json",
	}
	body, err := json.Marshal(in)
	require.NoError(t, err)
	response, astring := testRequest(t, jar, http.MethodPost, "/api/shorten/batch", string(body), ctype)
	defer response.Body.Close()
	require.Equal(t, http.StatusCreated, response.StatusCode)
	var answer []output
	require.NoError(t, err)
	err = json.Unmarshal([]byte(astring), &answer)
	require.NoError(t, err)
	require.Equal(t, len(in), len(answer))
}

func Test_Conflict(t *testing.T) {
	jar := newJar(t)
	t.Run("Test default Get handler", func(t *testing.T) {
		response, _ := testRequest(t, jar, http.MethodGet, "/", "", map[string]string{"Content-Type": "text/plain; charset=utf-8"})
		defer response.Body.Close()
		ctype := map[string]string{
			"Content-Type": "text/plain; charset=utf-8",
		}
		s := "http://kiuerhv9unvr.org"
		response, body1 := testRequest(t, jar, http.MethodPost, "/", s, ctype)
		defer response.Body.Close()
		matched, err := regexp.Match(`http:\/\/\w+\.\w+\.\w\.\w:\d+\/\w{8}`, []byte(body1))
		require.NoError(t, err)
		require.True(t, matched)
		time.Sleep(3 * time.Second)
		response, body2 := testRequest(t, jar, http.MethodPost, "/", s, ctype)
		defer response.Body.Close()
		require.Equal(t, http.StatusConflict, response.StatusCode)
		require.Equal(t, body1, body2)

	})
}

func Test_APIConflict(t *testing.T) {
	jar := newJar(t)
	t.Run("Test default Get handler", func(t *testing.T) {
		response, _ := testRequest(t, jar, http.MethodGet, "/", "", map[string]string{"Content-Type": "text/plain; charset=utf-8"})
		defer response.Body.Close()
		ctype := map[string]string{
			"Content-Type": "application/json",
		}
		type req struct {
			URL string `json:"url"` //{"url":"<some_url>"}
		}
		s := req{URL: "http://kiuerhv9unvr.org"}
		b, err := json.Marshal(s)
		require.NoError(t, err)
		response, body1 := testRequest(t, jar, http.MethodPost, "/api/shorten", string(b), ctype)
		defer response.Body.Close()
		matched, err := regexp.Match(`{"result":\"http:\/\/\w+\.\w+\.\w\.\w:\d+\/\w{8}\"}`, []byte(body1))
		require.NoError(t, err)
		require.True(t, matched)
		time.Sleep(3 * time.Second)
		response, body2 := testRequest(t, jar, http.MethodPost, "/api/shorten", string(b), ctype)
		defer response.Body.Close()
		require.Equal(t, http.StatusConflict, response.StatusCode)
		require.Equal(t, body1, body2)

	})
}

func Test_CmdWait(t *testing.T) {
	err := cmd.Wait()
	require.NoError(t, err)
	err = os.Remove(cmd.Path)
	require.NoError(t, err)
}
