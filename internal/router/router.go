package router

import (
	"context"
	"io"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/h1067675/shortUrl/cmd/configsurl"
	"github.com/h1067675/shortUrl/internal/compress"
	"github.com/h1067675/shortUrl/internal/logger"
	"go.uber.org/zap"
	"golang.org/x/crypto/acme/autocert"
)

const (
	keyUserID key = iota
	keyNewUser
)

// структуры
type (
	// Handler связывает с хэндлерами
	Applicator interface {
		Authorization(cookie string, hasToken bool) (token string, userid int, err error)
		CheckDB() (statuscode int)
		Shorten(url string, userid int) (body string, statusCode int)
		ShortenJSON(js []byte, userid int) (body []byte, statusCode int)
		ShortenBatchJSON(js []byte, userid int) (body []byte, statusCode int)
		Expand(shortCode string, userid int) (basedURL string, statusCode int)
		ExpandUserURLS(userid int, newuser bool) (body []byte, statusCode int)
		DeleteUserURLS(js []byte, userid int) (statusCode int)
		GetServerStats(ip net.IP) (body []byte, statusCode int)
	}
	// Router отвечает за маршрутизацию
	Router struct {
		Server *http.Server
		App    Applicator
		chi.Router
	}

	// key необходим для передачи через context
	key int
)

// // NewConnect создает коннектор
// func (r *Router) New() {
// 	r = &Router{
// 		chi.NewRouter(),
// 	}
// }

// RouterFunc делает маршрутизацию к хандлерам.
func (r *Router) RouterFunc(app Applicator) chi.Router {
	// Создаем chi роутер
	r.Router = chi.NewRouter()

	// Добавляем все функции middleware
	r.Router.Use(r.AuthorizationHandler)
	r.Router.Use(compress.CompressHandle)
	r.Router.Use(logger.RequestLogger)

	// Делаем маршрутизацию
	r.Router.Route("/", func(c chi.Router) {
		c.Post("/", r.ShortenHandler) // POST запрос отправляем на сокращение ссылки
		c.Route("/{id}", func(c chi.Router) {
			c.Get("/", r.ExpandHandler) // GET запрос с id направляем на извлечение ссылки
		})
		c.Route("/api/shorten", func(c chi.Router) {
			c.Post("/", r.ShortenJSONHandler)           // POST запрос с JSON телом
			c.Post("/batch", r.ShortenBatchJSONHandler) // POST запрос с множественным JSON телом
		})
		c.Route("/api/user", func(c chi.Router) {
			c.Get("/urls", r.ExpandUserURLSHandler)    // GET запрос на выдачу всех сокращенных ссылок пользователем
			c.Delete("/urls", r.DeleteUserURLSHandler) // DELETE запрос удаляет ссылки перечисленные в запросе
		})
		c.Route("/ping", func(c chi.Router) {
			c.Get("/", r.CheckDBHandler) // GET запрос проверяет работоспособность базы данных
		})
		c.Route("/api/internal/stats", func(c chi.Router) {
			c.Get("/", r.GetServerStatsHandler) // GET отдает статистику сервера
		})
	})

	// Добавляем связь с бизнес логикой
	r.App = app

	return r.Router
}

func New() Router {
	var r Router
	return r
}

func (r *Router) NewServerHTTP(conf configsurl.Config, app Applicator) *http.Server {
	// Определяем HTTP сервер и указываем адрес и ручку
	r.Server = &http.Server{
		Addr:    conf.GetConfig().ServerAddress,
		Handler: r.RouterFunc(app),
	}

	return r.Server
}

func (r *Router) StartServerHTTP(conf configsurl.Config) error {
	var err error
	// Определяем порядок работы сервера через HTTPS или HTTP.
	// Если определено в настройках использование HTTPS то запускаем сервер через ListenAndServeTLS
	if conf.EnableHTTPS.On {
		// конструируем менеджер TLS-сертификатов
		manager := &autocert.Manager{
			// директория для хранения сертификатов
			Cache: autocert.DirCache("cache-dir"),
			// функция, принимающая Terms of Service издателя сертификатов
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(conf.GetConfig().ServerAddress),
		}
		r.Server.TLSConfig = manager.TLSConfig()
		// Запускаем сервер с HTTPS
		err = r.Server.ListenAndServeTLS("", "")
	} else {
		// Запускаем сервер через HTTP
		err = r.Server.ListenAndServe()
	}

	if err != nil {
		logger.Log.Debug(err.Error(), zap.String("server address", conf.GetConfig().ServerAddress))
	}

	return err
}

// Authorization mildware осуществляет авторизацию пользователя.
func (r *Router) AuthorizationHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		var (
			err    error
			token  string
			userid int
			cookie *http.Cookie
			ctx    context.Context
		)
		logger.Log.Debug("Handler Authorization")
		cookie, err = request.Cookie("token")
		if err != nil {
			token, userid, err = r.App.Authorization("", false)
			if err == nil {
				cookie := &http.Cookie{
					Name:   "token",
					Value:  token,
					MaxAge: 60 * 60 * 24,
					Path:   "/",
				}
				http.SetCookie(response, cookie)
				ctx = context.WithValue(request.Context(), keyUserID, userid)
				ctx = context.WithValue(ctx, keyNewUser, true)
			} else {
				response.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			_, userid, err = r.App.Authorization(cookie.Value, true)
			if err == nil {
				ctx = context.WithValue(request.Context(), keyUserID, userid)
			}
		}

		next.ServeHTTP(response, request.WithContext(ctx))
	})
}

// CheckDBHandler проверяет наличие базы данных.
func (r *Router) CheckDBHandler(responce http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodGet {
		responce.WriteHeader(r.App.CheckDB())
		return
	}
	responce.WriteHeader(http.StatusInternalServerError)
}

// ShortenHandler сокращает URL полученные в теле запроса POST, принимает text/plain, проверят Content-type, присваивает правильный Content-type ответу,
// записывает правильный статус в ответ, получает тело запроса и если оно не пустое, то запрашивает сокращенную ссылку
// и возвращает ответ. Во всех иных случаях возвращает в ответе Bad request.
func (r *Router) ShortenHandler(responce http.ResponseWriter, request *http.Request) {
	logger.Log.Debug("Handler ShortenHandler")
	var err error
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
		body, statusCode := r.App.Shorten(string(url), request.Context().Value(keyUserID).(int))

		responce.WriteHeader(statusCode)
		responce.Write([]byte(body))
		return
	}
	responce.WriteHeader(http.StatusBadRequest)
}

// ShortenJSONHandler сокращает URL полученный в JSON, принимает application/json, проверят Content-type, присваивает правильный Content-type ответу,
// записывает правильный статус в ответ, получает тело запроса и если оно не пустое, то запрашивает сокращенную ссылку
// и возвращает ответ. Во всех иных случаях возвращает в ответе Bad request.
func (r *Router) ShortenJSONHandler(responce http.ResponseWriter, request *http.Request) {
	logger.Log.Debug("Handler ShortenJSONHandler")
	var err error
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
		body, statusCode := r.App.ShortenJSON(js, request.Context().Value(keyUserID).(int))
		responce.WriteHeader(statusCode)
		responce.Write(body)
		return
	}
	responce.WriteHeader(http.StatusBadRequest)
}

// ShortenBatchJSONHandler сокращает URL полученный в JSON, принимает application/json, проверят Content-type, присваивает правильный Content-type ответу,
// записывает правильный статус в ответ, получает тело запроса и если оно не пустое, то запрашивает сокращенную ссылку
// и возвращает ответ. Во всех иных случаях возвращает в ответе Bad request.
func (r *Router) ShortenBatchJSONHandler(responce http.ResponseWriter, request *http.Request) {
	logger.Log.Debug("Handler ShortenBatchJSONHandler")
	var err error
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
		body, statusCode := r.App.ShortenBatchJSON(js, request.Context().Value(keyUserID).(int))
		responce.WriteHeader(statusCode)
		responce.Write(body)
		return
	}
	responce.WriteHeader(http.StatusBadRequest)
}

// ExpandHandler получет адрес по короткой ссылке из GET запроса.
func (r *Router) ExpandHandler(responce http.ResponseWriter, request *http.Request) {
	logger.Log.Debug("Handler ExpandHandler")
	if request.Method == http.MethodGet {
		outURL, statusCode := r.App.Expand(request.URL.Path, request.Context().Value(keyUserID).(int))
		if outURL != "" {
			logger.Log.Debug("URL expanded " + outURL)
			responce.Header().Add("Location", outURL)
		}
		responce.WriteHeader(statusCode)
		return
	}
	responce.WriteHeader(http.StatusBadRequest)
}

// ExpandUserURLSHandler получает весь список сокращенных адресов пользователем прошедшим авторизацию.
func (r *Router) ExpandUserURLSHandler(responce http.ResponseWriter, request *http.Request) {
	logger.Log.Debug("Handler ExpandUserURLSHandler")
	ctx := request.Context()
	if request.Method == http.MethodGet {
		body, statusCode := r.App.ExpandUserURLS(ctx.Value(keyUserID).(int), ctx.Value(keyNewUser) == true)
		responce.Header().Add("Content-Type", "application/json")
		responce.WriteHeader(statusCode)
		responce.Write(body)
		logger.Log.Debug("take body to user urls", zap.String("body", string(body)))
		return
	}
	responce.WriteHeader(http.StatusBadRequest)
}

// DeleteUserURLSHandler удалет указанные в JSON сокращенные адреса пользователя прошедшего авторизацию.
func (r *Router) DeleteUserURLSHandler(responce http.ResponseWriter, request *http.Request) {
	logger.Log.Debug("Handler DeleteUserURLSHandler")
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
			statusCode := r.App.DeleteUserURLS(js, request.Context().Value(keyUserID).(int))
			responce.WriteHeader(statusCode)
			return
		}
	}
	responce.WriteHeader(http.StatusBadRequest)
}

// GetServerStats показывает статистику сервера если X-Real-IP входит в доверенную подсеть.
func (r *Router) GetServerStatsHandler(responce http.ResponseWriter, request *http.Request) {
	logger.Log.Debug("Handler GetServerStats")

	// берем ip пользователя из заголовка X-Real-IP и проверяем на вхождение его в доверенную подсеть
	ip := net.ParseIP(request.Header.Get("X-Real-IP"))
	body, statusCode := r.App.GetServerStats(ip)

	// Формируем ответ сервера
	responce.WriteHeader(statusCode)
	responce.Write(body)
}
