package main

import (
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
				response:    "http://localhost:8080/x",
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
			request := httptest.NewRequest(test.method, "/", nil)
			request.Header.Add("Content-Type", test.contentType)
			// создаём новый Recorder
			w := httptest.NewRecorder()
			shorten(w, request)

			res := w.Result()
			// проверяем код ответа
			assert.Equal(t, test.want.code, res.StatusCode)
			// получаем и проверяем тело запроса
			defer res.Body.Close()
			_, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}

func Test_expand(t *testing.T) {
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
			name:        "test expand #1",
			method:      http.MethodGet,
			contentType: "text/plain",
			body:        "",
			want: want{
				code:        http.StatusTemporaryRedirect,
				response:    "",
				contentType: "text/plain",
			},
		},
		{
			name:        "test shorten #2",
			method:      http.MethodPost,
			contentType: "text/plain",
			body:        "http://ya.ru/",
			want: want{
				code:        200,
				response:    "http://localhost:8080/x",
				contentType: "text/plain",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, "/", nil)
			// создаём новый Recorder
			w := httptest.NewRecorder()
			shorten(w, request)

			res := w.Result()
			// проверяем код ответа
			assert.Equal(t, test.want.code, res.StatusCode)
			// получаем и проверяем тело запроса
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)

			require.NoError(t, err)
			assert.Equal(t, test.want.response, resBody)
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}
