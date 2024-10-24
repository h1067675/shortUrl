package netservice

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/h1067675/shortUrl/internal/logger"
	"go.uber.org/zap"
)

type MemStorager interface {
	CreateShortURL(url string, adr string) string
	GetURL(url string) (l string, e error)
}

type Configurer interface {
	GetConfig() struct {
		ServerAddress string
		OuterAddress  string
	}
}

type Connect struct {
	Router  chi.Router
	Storage MemStorager
	Config  Configurer
}

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
//
// Добавить:
// 1. Валидацию на првильность указания ссылки которую нужно сократить
func (c *Connect) ShortenHandler(responce http.ResponseWriter, request *http.Request) {
	// проверяем на content-type
	if strings.Contains(request.Header.Get("Content-Type"), "text/plain") {
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
		// если тело запроса не пустое, то создаем сокращенный url и выводим в тело ответа
		if len(url) > 0 {
			body = c.Storage.CreateShortURL(string(url), c.Config.GetConfig().OuterAddress)
			responce.Write([]byte(body))
		}
		return
	}
	responce.WriteHeader(http.StatusBadRequest)
}

type JsRequest struct {
	URL string `json:"url"`
}
type JsResponce struct {
	URL string `json:"result"`
}

// ShortenJSONHandler - хандлер сокращения URL, юпринимает application/json, проверят Content-type, присваивает правильный Content-type ответу,
// записывает правильный статус в ответ, получает тело запроса и если оно не пустое, то запрашивает сокращенную ссылку
// и возвращает ответ. Во всех иных случаях возвращает в ответе Bad request
func (c *Connect) ShortenJSONHandler(responce http.ResponseWriter, request *http.Request) {
	// проверяем на content-type
	if strings.Contains(request.Header.Get("Content-Type"), "application/json") {
		// если прошли то присваиваем значение content-type: "application/json" и статус 201
		responce.Header().Add("Content-Type", "application/json")
		responce.WriteHeader(http.StatusCreated)
		// получаем тело запроса
		js, err := io.ReadAll(request.Body)
		if err != nil {
			log.Fatal(err)
			responce.WriteHeader(http.StatusBadRequest)
			return
		}
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

// expandHundler - хандлер получения адреса по короткой ссылке. Получаем короткую ссылку из GET запроса
func (c *Connect) ExpandHandler(responce http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodGet {
		outURL, err := c.Storage.GetURL("http://" + c.Config.GetConfig().ServerAddress + request.URL.Path)
		if err != nil {
			log.Fatal(err)
			responce.WriteHeader(http.StatusBadRequest)
		}
		responce.Header().Add("Location", outURL)
		responce.WriteHeader(http.StatusTemporaryRedirect)
	}
	responce.WriteHeader(http.StatusBadRequest)
}

// routerFunc - создает роутер chi и делает маршрутизацию к хандлерам
func (c *Connect) RouterFunc() chi.Router {
	c.Router = chi.NewRouter()
	// Делаем маршрутизацию
	c.Router.Route("/", func(r chi.Router) {
		r.Post("/", c.ShortenHandler) // POST запрос отправляем на сокращение ссылки
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", c.ExpandHandler) // GET запрос с id направляем на извлечение ссылки
		})
		r.Route("/api/shorten", func(r chi.Router) {
			r.Post("/", c.ShortenJSONHandler) // GET запрос с id направляем на извлечение ссылки
		})
	})
	logger.Log.Debug("Server is running", zap.String("server address", c.Config.GetConfig().ServerAddress))
	return c.Router
}

func (c *Connect) StartServer() {
	if err := http.ListenAndServe(c.Config.GetConfig().ServerAddress, logger.RequestLogger(c.RouterFunc())); err != nil {
		logger.Log.Fatal(err.Error(), zap.String("server address", c.Config.GetConfig().ServerAddress))
	}
}
