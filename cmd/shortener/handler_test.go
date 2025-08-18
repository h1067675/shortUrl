package main

// import (
// 	"bytes"
// 	"context"
// 	"errors"
// 	"io"
// 	"net"
// 	"net/http"
// 	"net/http/httptest"
// 	"strconv"
// 	"strings"
// 	"testing"

// 	"github.com/h1067675/shortUrl/cmd/storage"
// 	"github.com/h1067675/shortUrl/internal/logger"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// )

// type want struct {
// 	code        int
// 	response    string
// 	contentType string
// 	shortCode   string
// 	location    string
// }
// type test struct {
// 	name         string
// 	method       string
// 	methodAdd    string
// 	methodExpand string
// 	contentType  string
// 	body         string
// 	shortCode    string
// 	want         want
// }
// type TestStorage struct {
// 	storage.Storager
// 	InnerLinks  map[string]string
// 	OutterLinks map[string]string
// 	Users       map[int][]string
// 	UsersLinks  map[string][]int
// 	Test        test
// }

// func (s *TestStorage) CreateShortURL(url string, adr string, userid int) (string, error) {
// 	val, ok := s.OutterLinks[url]
// 	if ok {
// 		return val, nil
// 	}
// 	result := "http://" + adr + "/" + s.Test.shortCode
// 	s.OutterLinks[url] = result
// 	s.InnerLinks[result] = url
// 	return result, nil
// }
// func (s *TestStorage) GetURL(url string, userid int) (l string, e error) {
// 	l, ok := s.InnerLinks[url]
// 	if ok {
// 		return l, nil
// 	}
// 	return "", errors.New("link not found")
// }
// func (s *TestStorage) TakeTestData(test test) {
// 	s.Test = test
// }
// func (s *TestStorage) SaveToFile(file string) {

// }
// func (s *TestStorage) PingDB() bool {
// 	return false
// }
// func (s *TestStorage) GetNewUserID() (int, error) {
// 	return 1, nil
// }
// func (s *TestStorage) GetUserURLS(id int) (result []struct {
// 	ShortURL string
// 	URL      string
// }, err error) {
// 	return nil, nil
// }

// type NetAddressServer struct {
// 	Host string
// 	Port int
// }
// type Cnfg struct {
// 	NetAddressServerShortener NetAddressServer
// 	NetAddressServerExpand    NetAddressServer
// 	FileStoragePath           FilePath
// 	DatabasePath              DatabasePath
// }
// type FilePath struct {
// 	Path string
// }
// type DatabasePath struct {
// 	Path string
// }

// func (c *Cnfg) GetConfig() struct {
// 	ServerAddress   string
// 	OuterAddress    string
// 	FileStoragePath string
// 	DatabasePath    string
// } {
// 	return struct {
// 		ServerAddress   string
// 		OuterAddress    string
// 		FileStoragePath string
// 		DatabasePath    string
// 	}{ServerAddress: c.NetAddressServerShortener.String(), OuterAddress: c.NetAddressServerExpand.String(), FileStoragePath: c.FileStoragePath.Path, DatabasePath: c.DatabasePath.Path}
// }
// func (n *NetAddressServer) String() string {
// 	return n.Host + ":" + strconv.Itoa(n.Port)
// }
// func (n *NetAddressServer) Set(s string) (err error) {
// 	n.Host, n.Port, err = checkNetAddress(s, n.Host, n.Port)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func checkNetAddress(s string, h string, p int) (host string, port int, e error) {
// 	host, port = h, p
// 	v := strings.Split(s, "://")
// 	if len(v) < 1 || len(v) > 2 {
// 		e = errors.New("incorrect net address")
// 		return
// 	}
// 	if len(v) == 2 {
// 		s = v[1]
// 	}
// 	a := strings.Split(s, ":")
// 	if len(a) < 1 || len(a) > 2 {
// 		e = errors.New("incorrect net address")
// 		return
// 	}
// 	host = a[0]
// 	ip := net.ParseIP(host)
// 	if ip == nil && host != "localhost" {
// 		e = errors.New("incorrect net address")
// 		return
// 	}
// 	if a[1] != "" {
// 		port, e = strconv.Atoi(a[1])
// 		if e != nil || port < 0 || port > 65535 {
// 			e = errors.New("incorrect net address")
// 			return
// 		}
// 	}
// 	return
// }

// func Test_shortenHandler(t *testing.T) {
// 	tests := []test{{
// 		name:        "test shorten #1",
// 		method:      http.MethodPost,
// 		contentType: "text/plain",
// 		shortCode:   "",
// 		body:        "",
// 		want: want{
// 			code:        201,
// 			response:    "",
// 			contentType: "text/plain",
// 			shortCode:   "",
// 		},
// 	},
// 		{
// 			name:        "test shorten #2",
// 			method:      http.MethodPost,
// 			contentType: "text/plain; charset=utf-8",
// 			shortCode:   "12345678",
// 			body:        "http://ya.ru/",
// 			want: want{
// 				code:        201,
// 				response:    "http://ya.ru/",
// 				contentType: "text/plain",
// 				shortCode:   "12345678",
// 			},
// 		},
// 		{
// 			name:        "test shorten #3",
// 			method:      http.MethodPost,
// 			contentType: "text/json",
// 			shortCode:   "",
// 			body:        "http://ya.ru/",
// 			want: want{
// 				code:        400,
// 				response:    "",
// 				contentType: "",
// 				shortCode:   "",
// 			},
// 		},
// 	}

// 	logger.Initialize("debug")
// 	var strg = TestStorage{
// 		InnerLinks:  map[string]string{},
// 		OutterLinks: map[string]string{},
// 	}
// 	var cnf = Cnfg{
// 		NetAddressServerShortener: NetAddressServer{Host: "localhoxt", Port: 8080},
// 		NetAddressServerExpand:    NetAddressServer{Host: "localhoxt", Port: 8080},
// 	}
// 	var r = NewConnect(&strg, &cnf)

// 	for _, test := range tests {
// 		t.Run(test.name, func(t *testing.T) {
// 			strg.Test = test
// 			h := http.HandlerFunc(r.ShortenHandler)
// 			buf := bytes.NewBuffer([]byte(test.body))
// 			request, err := http.NewRequest(test.method, "/", buf)
// 			require.NoError(t, err)
// 			request.Header.Add("Content-Type", test.contentType)
// 			// rctx := chi.NewRouteContext()
// 			// request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))
// 			// rctx.URLParams.Add("name", "joe")
// 			w := httptest.NewRecorder()
// 			ctx := context.WithValue(request.Context(), keyUserID, 1)
// 			h.ServeHTTP(w, request.WithContext(ctx))
// 			resp := w.Result()
// 			defer resp.Body.Close()
// 			body, _ := io.ReadAll(resp.Body)
// 			var want string
// 			if test.want.shortCode != "" {
// 				want = "http://" + r.Config.GetConfig().OuterAddress + "/" + test.want.shortCode
// 			}
// 			assert.Equal(t, string(body), want)
// 			assert.Equal(t, test.want.code, resp.StatusCode)
// 			assert.Equal(t, test.want.contentType, resp.Header.Get("Content-Type"))

// 		})
// 	}
// }
// func Test_shortenJsonHandler(t *testing.T) {
// 	tests := []test{{
// 		name:        "test shorten #1",
// 		method:      http.MethodPost,
// 		contentType: "application/json",
// 		shortCode:   "",
// 		body:        "{}",
// 		want: want{
// 			code:        201,
// 			response:    "",
// 			contentType: "application/json",
// 			shortCode:   "",
// 		},
// 	},
// 		{
// 			name:        "test shorten #2",
// 			method:      http.MethodPost,
// 			contentType: "application/json; charset=utf-8",
// 			shortCode:   "12345678",
// 			body:        `{"url": "http://ya.ru/"}`,
// 			want: want{
// 				code:        201,
// 				response:    "http://ya.ru/",
// 				contentType: "application/json",
// 				shortCode:   "12345678",
// 			},
// 		},
// 		{
// 			name:        "test shorten #3",
// 			method:      http.MethodPost,
// 			contentType: "text/json",
// 			shortCode:   "",
// 			body:        "http://ya.ru/",
// 			want: want{
// 				code:        400,
// 				response:    "",
// 				contentType: "",
// 				shortCode:   "",
// 			},
// 		},
// 	}

// 	logger.Initialize("debug")
// 	var strg = TestStorage{
// 		InnerLinks:  map[string]string{},
// 		OutterLinks: map[string]string{},
// 	}
// 	var cnf = Cnfg{
// 		NetAddressServerShortener: NetAddressServer{Host: "localhoxt", Port: 8080},
// 		NetAddressServerExpand:    NetAddressServer{Host: "localhoxt", Port: 8080},
// 	}
// 	var r = NewConnect(&strg, &cnf)

// 	for _, test := range tests {
// 		t.Run(test.name, func(t *testing.T) {
// 			strg.Test = test
// 			h := http.HandlerFunc(r.ShortenJSONHandler)
// 			buf := bytes.NewBuffer([]byte(test.body))
// 			request, err := http.NewRequest(test.method, "/", buf)
// 			require.NoError(t, err)
// 			request.Header.Add("Content-Type", test.contentType)
// 			// rctx := chi.NewRouteContext()
// 			// request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))
// 			// rctx.URLParams.Add("name", "joe")
// 			w := httptest.NewRecorder()

// 			ctx := context.WithValue(request.Context(), keyUserID, 1)
// 			h.ServeHTTP(w, request.WithContext(ctx))
// 			resp := w.Result()
// 			defer resp.Body.Close()
// 			body, _ := io.ReadAll(resp.Body)
// 			var want string
// 			if test.want.shortCode != "" {
// 				want = `{"result":"http://` + r.Config.GetConfig().OuterAddress + "/" + test.want.shortCode + `"}`
// 			}
// 			assert.Equal(t, string(body), want)
// 			assert.Equal(t, test.want.code, resp.StatusCode)
// 			assert.Equal(t, test.want.contentType, resp.Header.Get("Content-Type"))

// 		})
// 	}
// }
// func Test_expand(t *testing.T) {
// 	tests := []test{
// 		{
// 			name:         "test expand #1",
// 			methodAdd:    http.MethodPost,
// 			methodExpand: http.MethodGet,
// 			contentType:  "text/plain",
// 			body:         "http://mail.ru",
// 			shortCode:    "12345678",
// 			want: want{
// 				code:        http.StatusTemporaryRedirect,
// 				response:    "",
// 				contentType: "text/plain",
// 				shortCode:   "12345678",
// 				location:    "http://mail.ru",
// 			},
// 		},
// 		{
// 			name:         "test expand #2",
// 			methodAdd:    http.MethodPost,
// 			methodExpand: http.MethodGet,
// 			contentType:  "text/plain; charset=utf-8",
// 			body:         "http://ya.ru/",
// 			shortCode:    "12345679",
// 			want: want{
// 				code:        http.StatusTemporaryRedirect,
// 				response:    "",
// 				contentType: "text/plain",
// 				shortCode:   "12345679",
// 				location:    "http://ya.ru/",
// 			},
// 		},
// 		{
// 			name:         "test expand #3",
// 			methodAdd:    http.MethodPost,
// 			methodExpand: http.MethodPost,
// 			contentType:  "text/plain; charset=utf-8",
// 			body:         "http://yandex.ru/",
// 			want: want{
// 				code:        http.StatusBadRequest,
// 				response:    "",
// 				contentType: "",
// 				location:    "",
// 			},
// 		},
// 	}

// 	logger.Initialize("debug")
// 	var strg = TestStorage{
// 		InnerLinks:  map[string]string{},
// 		OutterLinks: map[string]string{},
// 	}
// 	var cnf = Cnfg{
// 		NetAddressServerShortener: NetAddressServer{Host: "localhoxt", Port: 8080},
// 		NetAddressServerExpand:    NetAddressServer{Host: "localhoxt", Port: 8080},
// 	}
// 	var r = NewConnect(&strg, &cnf)

// 	for _, test := range tests {
// 		t.Run(test.name, func(t *testing.T) {
// 			strg.Test = test
// 			h := http.HandlerFunc(r.ShortenHandler)
// 			buf := bytes.NewBuffer([]byte(test.body))
// 			request, err := http.NewRequest(test.methodAdd, "/", buf)
// 			require.NoError(t, err)
// 			request.Header.Add("Content-Type", test.contentType)
// 			// rctx := chi.NewRouteContext()
// 			// request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))
// 			// rctx.URLParams.Add("name", "joe")
// 			w := httptest.NewRecorder()
// 			ctx := context.WithValue(request.Context(), keyUserID, 1)
// 			h.ServeHTTP(w, request.WithContext(ctx))
// 			resp := w.Result()
// 			defer resp.Body.Close()
// 			body, _ := io.ReadAll(resp.Body)
// 			h2 := http.HandlerFunc(r.ExpandHandler)
// 			request2, err := http.NewRequest(test.methodExpand, string(body), nil)
// 			require.NoError(t, err)
// 			request2.Header.Add("Content-Type", test.contentType)
// 			w2 := httptest.NewRecorder()
// 			h2.ServeHTTP(w2, request2.WithContext(ctx))
// 			resp2 := w2.Result()
// 			defer resp2.Body.Close()
// 			// проверяем код ответа
// 			assert.Equal(t, test.want.code, resp2.StatusCode)
// 			// ппроверяем location
// 			assert.Equal(t, test.want.location, resp2.Header.Get("Location"))
// 		})
// 	}
// }
// func Test_ShortenBatchJSONHandler(t *testing.T) {
// 	tests := []test{{
// 		name:        "test shorten #1",
// 		method:      http.MethodPost,
// 		contentType: "application/json",
// 		body:        `[{"correlation_id":"1","original_url": "http://ya.ru/"},{"correlation_id":"2","original_url": "http://yandex.ru/"}]`,
// 		want: want{
// 			code:        201,
// 			response:    `[{"correlation_id":"1","short_url": "*"},{"correlation_id":"2","short_url": "111"}]`,
// 			contentType: "application/json",
// 		},
// 	},
// 		{
// 			name:        "test shorten #1",
// 			method:      http.MethodPost,
// 			contentType: "application/json",
// 			body:        "{}",
// 			want: want{
// 				code:        201,
// 				response:    "{}",
// 				contentType: "application/json",
// 			},
// 		},
// 		{
// 			name:        "test shorten #1",
// 			method:      http.MethodPost,
// 			contentType: "application/json",
// 			body:        "",
// 			want: want{
// 				code:        201,
// 				response:    "",
// 				contentType: "application/json",
// 			},
// 		},
// 	}

// 	logger.Initialize("debug")
// 	var strg = TestStorage{
// 		InnerLinks:  map[string]string{},
// 		OutterLinks: map[string]string{},
// 	}
// 	var cnf = Cnfg{
// 		NetAddressServerShortener: NetAddressServer{Host: "localhoxt", Port: 8080},
// 		NetAddressServerExpand:    NetAddressServer{Host: "localhoxt", Port: 8080},
// 	}
// 	var r = NewConnect(&strg, &cnf)

// 	for _, test := range tests {
// 		t.Run(test.name, func(t *testing.T) {
// 			strg.Test = test
// 			h := http.HandlerFunc(r.ShortenBatchJSONHandler)
// 			buf := bytes.NewBuffer([]byte(test.body))
// 			request, err := http.NewRequest(test.method, "/", buf)
// 			require.NoError(t, err)
// 			request.Header.Add("Content-Type", test.contentType)
// 			// rctx := chi.NewRouteContext()
// 			// request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))
// 			// rctx.URLParams.Add("name", "joe")
// 			w := httptest.NewRecorder()

// 			ctx := context.WithValue(request.Context(), keyUserID, 1)
// 			h.ServeHTTP(w, request.WithContext(ctx))
// 			resp := w.Result()
// 			defer resp.Body.Close()
// 			_, err = io.ReadAll(resp.Body)
// 			assert.NoError(t, err)
// 			assert.Equal(t, test.want.code, resp.StatusCode)
// 			assert.Equal(t, test.want.contentType, resp.Header.Get("Content-Type"))

// 		})
// 	}
// }
