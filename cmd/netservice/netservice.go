package netservice

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"go.uber.org/zap"

	"github.com/h1067675/shortUrl/internal/compress"
	"github.com/h1067675/shortUrl/internal/logger"
)

// Интерфейс для Storage
type MemStorager interface {
	CreateShortURL(url string, adr string) string
	GetURL(url string) (l string, e error)
	SaveToFile(file string)
	PingDB() bool
}

// Интерфейс для Config
type Configurer interface {
	GetConfig() struct {
		ServerAddress   string
		OuterAddress    string
		FileStoragePath string
		DatabasePath    string
	}
}

// Структура с сетевыми методами
type Connect struct {
	Router  chi.Router
	Storage MemStorager
	Config  Configurer
}

// Функция создания коннектора
func NewConnect(i MemStorager, c Configurer) *Connect {
	var r = Connect{
		Router:  chi.NewRouter(),
		Storage: i,
		Config:  c,
	}
	return &r
}

// shortenHandler - хандлер сокращения URL, принимает text/plain, проверят Content-type, присваивает правильный Content-type ответу,
// записывает правильный статус в ответ, получает тело запроса и если оно не пустое, то запрашивает сокращенную ссылку
// и возвращает ответ. Во всех иных случаях возвращает в ответе Bad request
func (c *Connect) ShortenHandler(responce http.ResponseWriter, request *http.Request) {
	// проверяем на content-type
	if strings.Contains(request.Header.Get("Content-Type"), "text/plain") || strings.Contains(request.Header.Get("Content-type"), "application/x-gzip") {
		var body string
		// если прошли то присваиваем значение content-type: "text/plain" и статус 201
		responce.Header().Add("Content-Type", "text/plain")
		responce.WriteHeader(http.StatusCreated)
		// получаем тело запроса
		url, err := io.ReadAll(request.Body)
		if err != nil {
			log.Fatal(err)
			responce.WriteHeader(http.StatusBadRequest)
			return
		}
		logger.Log.Debug("Body", zap.String("type json", string(url)))
		// если тело запроса не пустое, то создаем сокращенный url и выводим в тело ответа
		if len(url) > 0 {
			body = c.Storage.CreateShortURL(string(url), c.Config.GetConfig().OuterAddress)
			responce.Write([]byte(body))
		}
		return
	}
	responce.WriteHeader(http.StatusBadRequest)
}

// Структура разбора batch json запроса
type JsBatchRequest struct {
	ID  string `json:"correlation_id"`
	URL string `json:"original_url"`
}

// Структура разбора batch json ответа
type JsBatchResponce struct {
	ID      string `json:"correlation_id"`
	SortURL string `json:"short_url"`
}

// Структура разбора json запроса
type JsRequest struct {
	URL string `json:"url"`
}

// Структура разбора json ответа
type JsResponce struct {
	URL string `json:"result"`
}

// ShortenJSONHandler - хандлер сокращения URL, юпринимает application/json, проверят Content-type, присваивает правильный Content-type ответу,
// записывает правильный статус в ответ, получает тело запроса и если оно не пустое, то запрашивает сокращенную ссылку
// и возвращает ответ. Во всех иных случаях возвращает в ответе Bad request
func (c *Connect) ShortenJSONHandler(responce http.ResponseWriter, request *http.Request) {
	// проверяем на content-type
	if strings.Contains(request.Header.Get("Content-Type"), "application/json") || strings.Contains(request.Header.Get("Content-type"), "application/x-gzip") {
		// если прошли то присваиваем значение content-type: "application/json" и статус 201
		responce.Header().Add("Content-Type", "application/json")
		responce.WriteHeader(http.StatusCreated)
		// получаем тело запроса
		js, err := io.ReadAll(request.Body)
		if err != nil {
			logger.Log.Error("Request wihtout body", zap.Error(err))
			responce.WriteHeader(http.StatusBadRequest)
			return
		}
		logger.Log.Debug("Body", zap.String("type json", string(js)))
		// если тело запроса не пустое, то создаем сокращенный url и выводим в тело ответа
		if len(js) > 0 {
			var url JsRequest
			if err := json.Unmarshal(js, &url); err != nil {
				logger.Log.Error("Error json parsing", zap.String("request body", string(js)))
			}
			if url.URL == "" {
				return
			}
			extURL := c.Storage.CreateShortURL(url.URL, c.Config.GetConfig().OuterAddress)
			result := JsResponce{URL: extURL}
			body, err := json.Marshal(result)
			if err != nil {
				logger.Log.Error("Error json serialization", zap.String("var", fmt.Sprint(result)))
			}
			responce.Write(body)
		}
		return
	}
	responce.WriteHeader(http.StatusBadRequest)
}

// ShortenBatchJSONHandler - хандлер сокращения URL, юпринимает application/json, проверят Content-type, присваивает правильный Content-type ответу,
// записывает правильный статус в ответ, получает тело запроса и если оно не пустое, то запрашивает сокращенную ссылку
// и возвращает ответ. Во всех иных случаях возвращает в ответе Bad request
func (c *Connect) ShortenBatchJSONHandler(responce http.ResponseWriter, request *http.Request) {
	// проверяем на content-type
	if strings.Contains(request.Header.Get("Content-Type"), "application/json") || strings.Contains(request.Header.Get("Content-type"), "application/x-gzip") {
		// если прошли то присваиваем значение content-type: "application/json" и статус 201
		responce.Header().Add("Content-Type", "application/json")
		responce.WriteHeader(http.StatusCreated)
		// получаем тело запроса
		js, err := io.ReadAll(request.Body)
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
				return
			}
			for _, e := range urls {
				resulturls = append(resulturls, JsBatchResponce{ID: e.ID, SortURL: c.Storage.CreateShortURL(e.URL, c.Config.GetConfig().OuterAddress)})
			}
			body, err := json.Marshal(resulturls)
			if err != nil {
				logger.Log.Error("Error json serialization")
			}
			responce.Write(body)
		}
		return
	}
	responce.WriteHeader(http.StatusBadRequest)
}

// expandHundler - хандлер получения адреса по короткой ссылке. Получаем короткую ссылку из GET запроса
func (c *Connect) ExpandHandler(responce http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodGet {
		outURL, err := c.Storage.GetURL("http://" + c.Config.GetConfig().ServerAddress + request.URL.Path)
		if err != nil {
			logger.Log.Error("Can't to get URL", zap.Error(err))
			responce.WriteHeader(http.StatusBadRequest)
		}
		responce.Header().Add("Location", outURL)
		responce.WriteHeader(http.StatusTemporaryRedirect)
		return
	}
	responce.WriteHeader(http.StatusBadRequest)
}

// CheckDBHandler -
func (c *Connect) CheckDBHandler(responce http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodGet && c.Storage.PingDB() {
		responce.WriteHeader(http.StatusOK)
	}
	responce.WriteHeader(http.StatusInternalServerError)
}

// routerFunc - создает роутер chi и делает маршрутизацию к хандлерам
func (c *Connect) RouterFunc() chi.Router {
	// Создаем chi роутер
	c.Router = chi.NewRouter()
	// Добавляем все функции middleware
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
			r.Post("/batch", c.ShortenBatchJSONHandler) // POST запрос с JSON телом
		})
		r.Route("/ping", func(r chi.Router) {
			r.Get("/", c.CheckDBHandler) // GET проверяет работоспособность базы данных
		})
	})
	logger.Log.Debug("Server is running", zap.String("server address", c.Config.GetConfig().ServerAddress))
	return c.Router
}

// Функция запуска сервера
func (c *Connect) StartServer() {
	if err := http.ListenAndServe(c.Config.GetConfig().ServerAddress, c.RouterFunc()); err != nil {
		logger.Log.Fatal(err.Error(), zap.String("server address", c.Config.GetConfig().ServerAddress))
	}
}
