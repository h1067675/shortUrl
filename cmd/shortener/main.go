package main

import (
	"errors"
	"flag"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strings"

	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi"
	"github.com/h1067675/shortUrl/cmd/configsurl"
)

func hasErrorFunc(s string) error {
	if s != "" {
		return errors.New(s)
	} else {
		return errors.New("has some problem")
	}
}

// randCharFunc - генерирует случайную букву латинского алфавита большую или маленькую или цифру
func randCharFunc() int {
	max := 122
	min := 48
	res := rand.Intn(max-min) + min
	if res > 57 && res < 65 || res > 90 && res < 97 {
		return randCharFunc()
	}
	return res
}

// createShortCodeFunc - генерирует новую короткую ссылку и проеряет на совпадение в "базе данных" если такая
// строка уже есть то делает рекурсию на саму себя пока не найдет уникальную ссылку
func createShortCodeFunc() string {
	shortURL := []byte("http://" + conf.NetAddressServerShortener.String() + "/")
	for i := 0; i < 8; i++ {
		shortURL = append(shortURL, byte(randCharFunc()))
	}
	result := string(shortURL)
	_, ok := shortUrls[result]
	if ok {
		return createShortCodeFunc()
	}
	return string(shortURL)
}

// createShortURLFunc - получает ссылку которую необходимо сократить и проверяет на наличие ее в "базе данных",
// если  есть, то возвращает уже готовый короткий URL, если нет то запрашивает новую случайную коротную ссылку
func createShortURLFunc(url string) string {
	val, ok := outUrls[url]
	if ok {
		return val
	}
	result := createShortCodeFunc()
	outUrls[url] = result
	shortUrls[result] = url
	return result
}

// getURLFunc - получает коротную ссылку и проверяет наличие ее в "базе данных" если существует, то возвращяет ее
// если нет, то возвращает ошибку
func getURLFunc(url string) (s string, e error) {
	s, ok := shortUrls[url]
	if ok {
		return s, nil
	}
	return "", hasErrorFunc("incorrect net address")
}

// checkHeaderFunc - проверяем заголовки полученные в hd на соответствие требуемому
func checkHeaderFunc(hd http.Header, key string, val string) bool {
	for k, v := range hd {
		if key == k {
			for _, vv := range v {
				if strings.Contains(vv, val) {
					return true
				}
			}
		}
	}
	return false
}

// shortenHandler - хандлер сокращения URL, проверят Content-type, присваивает правильный Content-type ответу,
// записывает правильный статус в ответ, получает тело запроса и если оно не пустое, то запрашивает сокращенную ссылку
// и возвращает ответ. Во всех иных случаях возвращает в ответе Bad request
//
// Добавить:
// 1. Валидацию на првильность указания ссылки которую нужно сократить
func shortenHandler(responce http.ResponseWriter, request *http.Request) {
	// проверяем на content-type
	if checkHeaderFunc(request.Header, "Content-Type", "text/plain") {
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
			body = createShortURLFunc(string(url))
			responce.Write([]byte(body))
		}
		return
	}
	responce.WriteHeader(http.StatusBadRequest)
}

// expandHundler - хандлер получения адреса по короткой ссылке. Получаем короткую ссылку из GET запроса
func expandHandler(responce http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodGet {
		outURL, err := getURLFunc("http://" + conf.NetAddressServerShortener.String() + request.URL.Path)
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
func routerFunc() chi.Router {
	r := chi.NewRouter()
	// Делаем маршрутизацию
	r.Route("/", func(r chi.Router) {
		r.Post("/", shortenHandler) // POST запрос отправляем на сокращение ссылки
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", expandHandler) // GET запрос с id направляем на извлечение ссылки
		})
	})
	return r
}

var outUrls = make(map[string]string)
var shortUrls = make(map[string]string)
var conf = configsurl.Config{
	NetAddressServerShortener: configsurl.NetAddressServer{Host: "localhost", Port: 8080},
	NetAddressServerExpand:    configsurl.NetAddressServer{Host: "localhost", Port: 8080},
}

type EnvConfig struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseURL       string `env:"BASE_URL"`
}

// parseFlagsFunc - разбираем строку атрибутов
func parseFlagsFunc() {
	flag.Var(&conf.NetAddressServerShortener, "a", "Net address shortener service (host:port)")
	flag.Var(&conf.NetAddressServerExpand, "b", "Net address expand service (host:port)")
	flag.Parse()
}

// envConfigFunc - забираем переменные окружения и если они установлены то указывам в конфиг из значения
func envConfigFunc() {
	var envConf EnvConfig
	err := env.Parse(&envConf)
	if err != nil {
		log.Fatal(err)
	}
	if envConf.ServerAddress != "" {
		conf.NetAddressServerShortener.Set(envConf.ServerAddress)
	}
	if envConf.BaseURL != "" {
		conf.NetAddressServerExpand.Set(envConf.BaseURL)
	}
}

func main() {
	parseFlagsFunc()
	envConfigFunc()
	log.Fatal(http.ListenAndServe(conf.NetAddressServerShortener.String(), routerFunc()))
}
