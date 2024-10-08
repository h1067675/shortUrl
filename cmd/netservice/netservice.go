package netservice

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi"

	"github.com/h1067675/shortUrl/cmd/configsurl"
	"github.com/h1067675/shortUrl/cmd/storage"
)

type Connect struct {
	Router chi.Router
	Base   storage.Storage
	Conf   configsurl.Config
}

// shortenHandler - хандлер сокращения URL, проверят Content-type, присваивает правильный Content-type ответу,
// записывает правильный статус в ответ, получает тело запроса и если оно не пустое, то запрашивает сокращенную ссылку
// и возвращает ответ. Во всех иных случаях возвращает в ответе Bad request
//
// Добавить:
// 1. Валидацию на првильность указания ссылки которую нужно сократить
func (c *Connect) ShortenHandler(responce http.ResponseWriter, request *http.Request) {
	// проверяем на content-type
	fmt.Printf("Метод: %s \n", request.Method)
	fmt.Printf("Context-type: %s \n", request.Header.Get("Content-Type"))
	fmt.Printf("Host: %s \n", request.Host)
	fmt.Printf("Адрес: %s \n", request.URL.Path)
	fmt.Printf("Адрес 2: %s \n", request.URL.Host)
	fmt.Println(request)
	fmt.Println(c)
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
			body = c.Base.CreateShortURL(string(url), c.Conf.NetAddressServerShortener.String())
			responce.Write([]byte(body))
		}
		return
	}
	responce.WriteHeader(http.StatusBadRequest)
}

// expandHundler - хандлер получения адреса по короткой ссылке. Получаем короткую ссылку из GET запроса
func (c *Connect) ExpandHandler(responce http.ResponseWriter, request *http.Request) {
	fmt.Printf("Метод: %s \n", request.Method)
	fmt.Printf("Context-type: %s \n", request.Header.Get("Content-Type"))
	fmt.Printf("Host: %s \n", request.Host)
	fmt.Printf("Адрес: %s \n", request.URL.Path)
	fmt.Printf("Адрес 2: %s \n", request.URL.Host)
	fmt.Println(request)
	fmt.Println(c)
	if request.Method == http.MethodGet {
		outURL, err := c.Base.GetURL("http://" + c.Conf.NetAddressServerShortener.String() + request.URL.Path)
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
	})
	return c.Router
}
