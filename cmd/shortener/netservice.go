package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"go.uber.org/zap"

	"github.com/h1067675/shortUrl/cmd/authorization"
	"github.com/h1067675/shortUrl/cmd/storage"
	"github.com/h1067675/shortUrl/internal/compress"
	"github.com/h1067675/shortUrl/internal/logger"
)

const (
	keyUserID key = iota
	keyNewUser
)

var ErrLinkExsist = errors.New("link already exsist")

type (
	key int

	// Configurer интерфейс для Config.
	Configurer interface {
		GetConfig() struct {
			ServerAddress   string
			OuterAddress    string
			FileStoragePath string
			DatabasePath    string
		}
	}
	// Connect определяет сетевые методы.
	Connect struct {
		Router  chi.Router
		Storage storage.Storager
		Config  Configurer
	}

	// JsBatchRequest определяет порядок разбора batch json запроса.
	JsBatchRequest struct {
		ID  string `json:"correlation_id"`
		URL string `json:"original_url"`
	}

	// JsBatchResponce определяет порядок разбора batch json ответа.
	JsBatchResponce struct {
		ID      string `json:"correlation_id"`
		SortURL string `json:"short_url"`
	}

	// JsRequest определяет порядок разбора json запроса.
	JsRequest struct {
		URL string `json:"url"`
	}

	// JsResponce определяет порядок разбора json ответа.
	JsResponce struct {
		URL string `json:"result"`
	}

	// JsUserRequest определяет порядок разбора json ответа с перечнем сокращенных ссылков.
	JsUserRequest struct {
		ShortURL    string `json:"short_url"`
		OriginalURL string `json:"original_url"`
	}
)

// NewConnect создает коннектор
func NewConnect(i storage.Storager, c Configurer) *Connect {
	var r = Connect{
		Router:  chi.NewRouter(),
		Storage: i,
		Config:  c,
	}
	return &r
}

// RouterFunc создает роутер chi и делает маршрутизацию к хандлерам.
func (c *Connect) RouterFunc() chi.Router {
	// Создаем chi роутер
	c.Router = chi.NewRouter()
	// Добавляем все функции middleware
	c.Router.Use(c.Authorization)
	c.Router.Use(compress.CompressHandle)
	c.Router.Use(logger.RequestLogger)

	// Делаем маршрутизацию
	c.Router.Route("/", func(r chi.Router) {
		r.Post("/", c.ShortenHandler) // POST запрос отправляем на сокращение ссылки
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", c.ExpandHandler) // GET запрос с id направляем на извлечение ссылки
		})
		r.Route("/api/shorten", func(r chi.Router) {
			r.Post("/", c.ShortenJSONHandler)           // POST запрос с JSON телом
			r.Post("/batch", c.ShortenBatchJSONHandler) // POST запрос с множественным JSON телом
		})
		r.Route("/api/user", func(r chi.Router) {
			r.Get("/urls", c.ExpandUserURLSHandler)    // GET запрос на выдачу всех сокращенных ссылок пользователем
			r.Delete("/urls", c.DeleteUserURLSHandler) // DELETE запрос удаляет ссылки перечисленные в запросе
		})
		r.Route("/ping", func(r chi.Router) {
			r.Get("/", c.CheckDBHandler) // GET запрос проверяет работоспособность базы данных
		})
	})
	logger.Log.Debug("Server is running", zap.String("server address", c.Config.GetConfig().ServerAddress))
	return c.Router
}

// StartServer запускает сервер.
func (c *Connect) StartServer() {
	if err := http.ListenAndServe(c.Config.GetConfig().ServerAddress, c.RouterFunc()); err != nil {
		logger.Log.Fatal(err.Error(), zap.String("server address", c.Config.GetConfig().ServerAddress))
	}
}

// CheckDBHandler проверяет наличие базы данных.
func (c *Connect) CheckDBHandler(responce http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodGet && c.Storage.PingDB() {
		responce.WriteHeader(http.StatusOK)
	}
	responce.WriteHeader(http.StatusInternalServerError)
}

// Authorization осуществляет авторизацию пользователя.
func (c *Connect) Authorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		var (
			err    error
			userid int
			cookie *http.Cookie
			ctx    context.Context
		)
		logger.Log.Debug("checking authorization")
		cookie, err = request.Cookie("token")
		if err == nil {
			logger.Log.Debug("user cookie", zap.String("cookie", cookie.Value))
			userid, err = authorization.CheckToken(cookie.Value)
			if err == nil {
				ctx = context.WithValue(request.Context(), keyUserID, userid)
			}
		} else {
			userid, err := c.Storage.GetNewUserID()
			logger.Log.Debug("new user", zap.Int("id", userid))
			if err != nil {
				logger.Log.Error("don't can to get new user ID", zap.Error(err))
			}
			token, err := authorization.SetToken(userid)
			if err != nil {
				logger.Log.Error("don't can to create token", zap.Error(err))
			}
			cookie := &http.Cookie{
				Name:   "token",
				Value:  token,
				MaxAge: 60 * 60 * 24,
				Path:   "/",
			}
			http.SetCookie(response, cookie)
			ctx = context.WithValue(request.Context(), keyUserID, userid)
			ctx = context.WithValue(ctx, keyNewUser, true)
		}
		next.ServeHTTP(response, request.WithContext(ctx))
	})
}

// ShortenHandler сокращает URL полученные в теле запроса POST, принимает text/plain, проверят Content-type, присваивает правильный Content-type ответу,
// записывает правильный статус в ответ, получает тело запроса и если оно не пустое, то запрашивает сокращенную ссылку
// и возвращает ответ. Во всех иных случаях возвращает в ответе Bad request.
func (c *Connect) ShortenHandler(responce http.ResponseWriter, request *http.Request) {
	var err error
	var body string
	// проверяем на content-type
	if strings.Contains(request.Header.Get("Content-Type"), "text/plain") || strings.Contains(request.Header.Get("Content-type"), "application/x-gzip") {
		// если прошли то присваиваем значение content-type: "text/plain" и статус 201
		responce.Header().Add("Content-Type", "text/plain")
		// получаем тело запроса
		var url []byte
		url, err = io.ReadAll(request.Body)
		if err != nil {
			log.Fatal(err)
			responce.WriteHeader(http.StatusBadRequest)
			return
		}
		logger.Log.Debug("Body", zap.String("request URL", string(url)))
		// если тело запроса не пустое, то создаем сокращенный url и выводим в тело ответа
		if len(url) > 0 {
			body, err = c.Storage.CreateShortURL(string(url), c.Config.GetConfig().OuterAddress, request.Context().Value(keyUserID).(int))
			if err != nil {
				responce.WriteHeader(http.StatusConflict)
			}
			logger.Log.Debug("Result body", zap.String("sort URL", string(body)))
			c.Storage.SaveToFile(c.Config.GetConfig().FileStoragePath)

		}
		responce.WriteHeader(http.StatusCreated)
		responce.Write([]byte(body))
		return
	}
	responce.WriteHeader(http.StatusBadRequest)
}

// ShortenJSONHandler сокращает URL полученный в JSON, принимает application/json, проверят Content-type, присваивает правильный Content-type ответу,
// записывает правильный статус в ответ, получает тело запроса и если оно не пустое, то запрашивает сокращенную ссылку
// и возвращает ответ. Во всех иных случаях возвращает в ответе Bad request.
func (c *Connect) ShortenJSONHandler(responce http.ResponseWriter, request *http.Request) {
	var err error
	var body []byte
	// проверяем на content-type
	if strings.Contains(request.Header.Get("Content-Type"), "application/json") || strings.Contains(request.Header.Get("Content-type"), "application/x-gzip") {
		// если прошли то присваиваем значение content-type: "application/json" и статус 201
		responce.Header().Add("Content-Type", "application/json")
		// получаем тело запроса
		var js []byte
		js, err = io.ReadAll(request.Body)
		if err != nil {
			logger.Log.Error("Request wihtout body", zap.Error(err))
			responce.WriteHeader(http.StatusBadRequest)
			return
		}
		logger.Log.Debug("Body", zap.String("type json", string(js)))
		// если тело запроса не пустое, то создаем сокращенный url и выводим в тело ответа
		if len(js) > 0 {
			var url JsRequest
			if err = json.Unmarshal(js, &url); err != nil {
				logger.Log.Error("Error json parsing", zap.String("request body", string(js)))
			}
			if url.URL == "" {
				responce.WriteHeader(http.StatusCreated)
				return
			}
			extURL, err := c.Storage.CreateShortURL(url.URL, c.Config.GetConfig().OuterAddress, request.Context().Value(keyUserID).(int))
			if err != nil {
				responce.WriteHeader(http.StatusConflict)
			}
			result := JsResponce{URL: extURL}
			body, err = json.Marshal(result)
			if err != nil {
				logger.Log.Error("Error json serialization", zap.String("var", fmt.Sprint(result)))
			}
			c.Storage.SaveToFile(c.Config.GetConfig().FileStoragePath)
		}
		responce.WriteHeader(http.StatusCreated)
		responce.Write(body)
		return
	}
	responce.WriteHeader(http.StatusBadRequest)
}

// ShortenBatchJSONHandler сокращает URL полученный в JSON, принимает application/json, проверят Content-type, присваивает правильный Content-type ответу,
// записывает правильный статус в ответ, получает тело запроса и если оно не пустое, то запрашивает сокращенную ссылку
// и возвращает ответ. Во всех иных случаях возвращает в ответе Bad request.
func (c *Connect) ShortenBatchJSONHandler(responce http.ResponseWriter, request *http.Request) {
	var err error
	var body []byte
	// проверяем на content-type
	if strings.Contains(request.Header.Get("Content-Type"), "application/json") || strings.Contains(request.Header.Get("Content-type"), "application/x-gzip") {
		// если прошли то присваиваем значение content-type: "application/json" и статус 201
		responce.Header().Add("Content-Type", "application/json")
		// получаем тело запроса
		var js []byte
		js, err = io.ReadAll(request.Body)
		if err != nil {
			logger.Log.Error("Request wihtout body", zap.Error(err))
			responce.WriteHeader(http.StatusBadRequest)
			return
		}
		logger.Log.Debug("Body", zap.String("type json", string(js)))
		// если тело запроса не пустое, то создаем сокращенный url и выводим в тело ответа
		// jjs := []byte(`[{"correlation_id": "1","original_url": "ya.ru"}]`)
		if len(js) > 0 {
			var urls []JsBatchRequest
			var resulturls []JsBatchResponce
			if err := json.Unmarshal(js, &urls); err != nil {
				logger.Log.Error("Error json parsing", zap.String("request body", string(js)))
			}
			if len(urls) == 0 {
				responce.WriteHeader(http.StatusCreated)
				return
			}
			for _, e := range urls {
				extURL, _ := c.Storage.CreateShortURL(e.URL, c.Config.GetConfig().OuterAddress, request.Context().Value(keyUserID).(int))
				resulturls = append(resulturls, JsBatchResponce{ID: e.ID, SortURL: extURL})
			}
			body, _ = json.Marshal(resulturls)
			c.Storage.SaveToFile(c.Config.GetConfig().FileStoragePath)
		}
		responce.WriteHeader(http.StatusCreated)
		responce.Write(body)
		return
	}
	responce.WriteHeader(http.StatusBadRequest)
}

// ExpandHandler получет адрес по короткой ссылке из GET запроса.
func (c *Connect) ExpandHandler(responce http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodGet {
		ctx := request.Context()
		outURL, err := c.Storage.GetURL("http://"+c.Config.GetConfig().ServerAddress+request.URL.Path, ctx.Value(keyUserID).(int))
		logger.Log.Debug("error from func", zap.Error(err))
		if err == storage.ErrLinkDeleted {
			logger.Log.Debug("URL has been deleted", zap.Error(err))
			responce.WriteHeader(http.StatusGone)
			return
		} else if err != nil {
			logger.Log.Debug("Can't to get URL", zap.Error(err))
			responce.WriteHeader(http.StatusBadRequest)
		}
		responce.Header().Add("Location", outURL)
		responce.WriteHeader(http.StatusTemporaryRedirect)
		return
	}
	responce.WriteHeader(http.StatusBadRequest)
}

// ExpandUserURLSHandler получает весь список сокращенных адресов пользователем прошедшив авторизацию.
func (c *Connect) ExpandUserURLSHandler(responce http.ResponseWriter, request *http.Request) {
	var urls []JsUserRequest
	ctx := request.Context()

	if request.Method == http.MethodGet {

		if ctx.Value(keyNewUser) == true {
			responce.WriteHeader(http.StatusUnauthorized)
			return
		}
		urlsr, _ := c.Storage.GetUserURLS(ctx.Value(keyUserID).(int))
		for _, e := range urlsr {
			urls = append(urls, JsUserRequest{ShortURL: e.ShortURL, OriginalURL: e.URL})
		}
		if len(urls) == 0 {
			responce.WriteHeader(http.StatusNoContent)
			return
		}
		body, err := json.Marshal(urls)
		if err != nil {
			logger.Log.Debug("can't serialized json answer")
		}
		responce.Header().Add("Content-Type", "application/json")
		responce.WriteHeader(http.StatusOK)
		responce.Write(body)
		logger.Log.Debug("take body to user urls", zap.String("body", string(body)))
		return
	}
	responce.WriteHeader(http.StatusBadRequest)
}

// DeleteUserURLSHandler удалет указанные в JSON сокращенные адреса пользователя прошедшего авторизацию.
func (c *Connect) DeleteUserURLSHandler(responce http.ResponseWriter, request *http.Request) {
	var err error
	if strings.Contains(request.Header.Get("Content-Type"), "application/json") || strings.Contains(request.Header.Get("Content-type"), "application/x-gzip") {
		var js []byte
		js, err = io.ReadAll(request.Body)
		if err != nil {
			logger.Log.Error("Request wihtout body", zap.Error(err))
			responce.WriteHeader(http.StatusBadRequest)
			return
		}
		logger.Log.Debug("Body", zap.String("type json", string(js)))
		if len(js) > 0 {
			var ids storage.DeleteUserURLS
			ctx := request.Context()
			ids.UserID = ctx.Value(keyUserID).(int)
			if err := json.Unmarshal(js, &ids.LinksIDS); err != nil {
				logger.Log.Error("Error json parsing", zap.String("request body", string(js)))
			}
			c.Storage.DeleteUserURLS(ids)
			responce.WriteHeader(http.StatusAccepted)
		}
	}
	responce.WriteHeader(http.StatusBadRequest)
}
