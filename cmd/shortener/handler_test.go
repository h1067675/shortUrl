package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_shorten(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name        string
		method      string
		contentType string
		body        string
		want        want
	}{
		{
			name:        "test shorten #1",
			method:      http.MethodPost,
			contentType: "text/plain",
			body:        "",
			want: want{
				code:        201,
				response:    "",
				contentType: "text/plain",
			},
		},
		{
			name:        "test shorten #2",
			method:      http.MethodPost,
			contentType: "text/plain; charset=utf-8",
			body:        "http://ya.ru/",
			want: want{
				code:        201,
				response:    "http://ya.ru/",
				contentType: "text/plain",
			},
		},
		{
			name:        "test shorten #3",
			method:      http.MethodPost,
			contentType: "text/json",
			body:        "http://ya.ru/",
			want: want{
				code:        400,
				response:    "",
				contentType: "",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buf := bytes.NewBuffer([]byte(test.body))
			request := httptest.NewRequest(test.method, "/", buf)

			request.Header.Add("Content-Type", test.contentType)
			// создаём новый Recorder
			w := httptest.NewRecorder()
			shorten(w, request)

			res := w.Result()
			// проверяем код ответа
			assert.Equal(t, test.want.code, res.StatusCode)
			// получаем и проверяем тело запроса
			defer res.Body.Close()
			body, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			assert.Equal(t, outUrls[test.want.response], string(body))
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}

func Test_expand(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
		location    string
	}
	tests := []struct {
		name         string
		methodAdd    string
		methodExpand string
		contentType  string
		body         string
		want         want
	}{
		{
			name:         "test expand #1",
			methodAdd:    http.MethodPost,
			methodExpand: http.MethodGet,
			contentType:  "text/plain",
			body:         "http://mail.ru",
			want: want{
				code:        http.StatusTemporaryRedirect,
				response:    "",
				contentType: "text/plain",
				location:    "http://mail.ru",
			},
		},
		{
			name:         "test expand #2",
			methodAdd:    http.MethodPost,
			methodExpand: http.MethodGet,
			contentType:  "text/plain; charset=utf-8",
			body:         "http://ya.ru/",
			want: want{
				code:        http.StatusTemporaryRedirect,
				response:    "",
				contentType: "text/plain",
				location:    "http://ya.ru/",
			},
		},
		{
			name:         "test expand #3",
			methodAdd:    http.MethodPost,
			methodExpand: http.MethodPost,
			contentType:  "text/plain; charset=utf-8",
			body:         "http://yandex.ru/",
			want: want{
				code:        http.StatusBadRequest,
				response:    "",
				contentType: "",
				location:    "",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buf := bytes.NewBuffer([]byte(test.body))
			request := httptest.NewRequest(test.methodAdd, "/", buf)
			request.Header.Add("Content-Type", test.contentType)
			// создаём новый Recorder
			w := httptest.NewRecorder()
			shorten(w, request)
			res := w.Result()
			// получаем и проверяем тело запроса
			defer res.Body.Close()
			body, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			buf2 := bytes.NewBuffer([]byte(test.body))
			request2 := httptest.NewRequest(test.methodExpand, string(body), buf2)
			request2.Header.Add("Content-Type", test.contentType)
			// создаём новый Recorder
			w2 := httptest.NewRecorder()
			expand(w2, request2)

			res2 := w2.Result()
			// проверяем код ответа
			assert.Equal(t, test.want.code, res2.StatusCode)
			// получаем и проверяем тело запроса
			defer res2.Body.Close()
			_, err2 := io.ReadAll(res2.Body)
			require.NoError(t, err2)

			assert.Equal(t, test.want.location, res2.Header.Get("Location"))
		})
	}
}
