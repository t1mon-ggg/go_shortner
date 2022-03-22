package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testRequest(t *testing.T, ts *httptest.Server, method, path, body string) (*http.Response, string) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	bodyreq := strings.NewReader(body)
	req, err := http.NewRequest(method, ts.URL+path, bodyreq)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)

	respBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	defer resp.Body.Close()

	return resp, string(respBody)
}

func TestRouter(t *testing.T) {
	db := make(tmpDB)
	db["ABCDabcd"] = "http://example.org"
	type want struct {
		statusCode int
		data       string
	}
	type request struct {
		method string
		query  string
		body   string
		rtype  string
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
			},
			want: want{
				statusCode: 201,
				data:       ":8080/\\w{8}",
			},
		},
		{
			name: "Unshort static url test",
			request: request{
				method: http.MethodGet,
				query:  "/ABCDabcd",
				body:   "http://example.org",
				rtype:  "GetLong",
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
			},
			want: want{
				statusCode: 400,
				data:       "",
			},
		},
	}
	for _, tt := range tests {
		//log.Println(db)
		r := chi.NewRouter()
		r.Route("/", db.Router)
		ts := httptest.NewServer(r)
		defer ts.Close()
		t.Run(tt.name, func(t *testing.T) {
			response, body := testRequest(t, ts, tt.request.method, tt.request.query, tt.request.body)
			switch tt.request.rtype {
			case "CreateShort":
				require.Equal(t, tt.want.statusCode, response.StatusCode)
				matched, err := regexp.Match("\\w{8}", []byte(body))
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
				t.Log("!!!!!!!!!!!!!!!!!!!!!!!!!")
				rex := regexp.MustCompile("\\w{8}")
				short := "/" + rex.FindString(body)
				step2, _ := testRequest(t, ts, http.MethodGet, short, "")
				assert.Equal(t, tt.want.statusCode, step2.StatusCode)
				if tt.want.statusCode != 400 {
					header := step2.Header.Get("Location")
					require.Equal(t, tt.want.data, header)
				}
			default:
				require.Equal(t, tt.want.statusCode, response.StatusCode)
			}
		})
	}
}
