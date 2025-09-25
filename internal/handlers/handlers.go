package handlers

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"io"
// 	"log"
// 	"net"
// 	"net/http"
// 	"os"
// 	"os/signal"
// 	"strings"
// 	"syscall"
// 	"time"

// 	"go.uber.org/zap"
// 	"golang.org/x/crypto/acme/autocert"

// 	"github.com/h1067675/shortUrl/cmd/authorization"
// 	"github.com/h1067675/shortUrl/cmd/configsurl"
// 	"github.com/h1067675/shortUrl/cmd/storage"
// 	"github.com/h1067675/shortUrl/internal/logger"
// 	"github.com/h1067675/shortUrl/internal/router"
// )

// const (
// 	keyUserID key = iota
// 	keyNewUser
// )

// // структуры
// type (
// 	// key необходим для передачи через context
// 	key int

// 	// Application описывает структуру зависимостей для доступа к базе данных и настройкам приложения
// 	Application struct {
// 		Storage *storage.Storage
// 		Config  *configsurl.Config
// 		Router  router.Router
// 	}

// 	// JsBatchRequest определяет порядок разбора batch json запроса.
// 	JsBatchRequest struct {
// 		ID  string `json:"correlation_id"`
// 		URL string `json:"original_url"`
// 	}

// 	// JsBatchResponce определяет порядок разбора batch json ответа.
// 	JsBatchResponce struct {
// 		ID      string `json:"correlation_id"`
// 		SortURL string `json:"short_url"`
// 	}

// 	// JsRequest определяет порядок разбора json запроса.
// 	JsRequest struct {
// 		URL string `json:"url"`
// 	}

// 	// JsResponce определяет порядок разбора json ответа.
// 	JsResponce struct {
// 		URL string `json:"result"`
// 	}

// 	// JsUserRequest определяет порядок разбора json ответа с перечнем сокращенных ссылков.
// 	JsUserRequest struct {
// 		ShortURL    string `json:"short_url"`
// 		OriginalURL string `json:"original_url"`
// 	}

// 	// JsStat определяет фотмат выдачи статистики сервера
// 	JsStat struct {
// 		Users int `json:"users"`
// 		URLS  int `json:"urls"`
// 	}
// )

// // New задает зависимости между пакетами
// func (app *Application) New(s *storage.Storage, c *configsurl.Config, r router.Router) {
// 	app.Storage = s
// 	app.Config = c
// 	app.Router = r
// }

// // StartServer запускает сервер.
// func (app *Application) StartServer() {
// 	var err error

// 	// Определяем сервер и указываем адрес и ручку
// 	server := &http.Server{
// 		Addr:    app.Config.GetConfig().ServerAddress,
// 		Handler: app.Router.RouterFunc(app),
// 	}

// 	// Запускаем ожидание сигналов на завершение работы
// 	idleConnsClosed := app.waitSysSignals(server)

// 	// Определяем порядок работы сервера через HTTPS или HHTP.
// 	// Если определено в настройках использование HTTPS то запускаем сервер через ListenAndServeTLS
// 	if app.Config.EnableHTTPS.On {
// 		// конструируем менеджер TLS-сертификатов
// 		manager := &autocert.Manager{
// 			// директория для хранения сертификатов
// 			Cache: autocert.DirCache("cache-dir"),
// 			// функция, принимающая Terms of Service издателя сертификатов
// 			Prompt:     autocert.AcceptTOS,
// 			HostPolicy: autocert.HostWhitelist(app.Config.GetConfig().ServerAddress),
// 		}
// 		server.TLSConfig = manager.TLSConfig()
// 		// Запускаем сервер с HTTPS
// 		err = server.ListenAndServeTLS("", "")
// 	} else {
// 		// Запускаем сервер через HTTP
// 		err = server.ListenAndServe()
// 	}
// 	if err != nil {
// 		logger.Log.Debug(err.Error(), zap.String("server address", app.Config.GetConfig().ServerAddress))
// 	}

// 	// Ждем мягкого завершения работы сервера
// 	<-idleConnsClosed

// 	// Сохраняем хранилище из мапы в файл
// 	app.Storage.SaveToFile(app.Config.GetConfig().FileStoragePath)
// 	logger.Log.Info("Storage saved to " + app.Config.GetConfig().FileStoragePath)

// 	// Сообщаем об окончании работы сервера
// 	logger.Log.Info("Server has shutdown in graceful mode")
// }

// // WaitSysSignals определяет логику для отслеживания сигналов завершения работы приложения и
// // реализует процесс мягкого завершения работы приложения.
// func (app *Application) waitSysSignals(server *http.Server) chan struct{} {
// 	// через этот канал сообщим основному потоку, что соединения закрыты
// 	idleConnsClosed := make(chan struct{})
// 	// канал для перенаправления прерываний
// 	sigint := make(chan os.Signal, 1)
// 	// регистрируем перенаправление прерываний
// 	signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
// 	// запускаем горутину обработки пойманных прерываний
// 	go func() {
// 		// читаем из канала прерываний
// 		<-sigint

// 		// создаем контекст с таймаутом на завершение операций сервером
// 		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
// 		defer cancel()

// 		// получили сигнал, запускаем процедуру graceful shutdown сервера
// 		if err := server.Shutdown(ctx); err != nil {
// 			logger.Log.Debug(err.Error(), zap.String("server address", app.Config.GetConfig().ServerAddress))
// 		}

// 		// сообщаем основному потоку, что все сетевые соединения обработаны и закрыты
// 		close(idleConnsClosed)
// 	}()
// 	return idleConnsClosed
// }

// // Authorization mildware осуществляет авторизацию пользователя.
// func (app *Application) Authorization(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
// 		logger.Log.Debug("Handler Authorization")
// 		var (
// 			err    error
// 			userid int
// 			cookie *http.Cookie
// 			ctx    context.Context
// 		)
// 		logger.Log.Debug("checking authorization")
// 		cookie, err = request.Cookie("token")
// 		if err == nil {
// 			logger.Log.Debug("user cookie", zap.String("cookie", cookie.Value))
// 			userid, err = authorization.CheckToken(cookie.Value)
// 			if err == nil {
// 				ctx = context.WithValue(request.Context(), keyUserID, userid)
// 			}
// 		} else {
// 			userid, err := app.Storage.GetNewUserID()
// 			logger.Log.Debug("new user", zap.Int("id", userid))
// 			if err != nil {
// 				logger.Log.Error("don't can to get new user ID", zap.Error(err))
// 			}
// 			token, err := authorization.SetToken(userid)
// 			if err != nil {
// 				logger.Log.Error("don't can to create token", zap.Error(err))
// 			}
// 			cookie := &http.Cookie{
// 				Name:   "token",
// 				Value:  token,
// 				MaxAge: 60 * 60 * 24,
// 				Path:   "/",
// 			}
// 			http.SetCookie(response, cookie)
// 			ctx = context.WithValue(request.Context(), keyUserID, userid)
// 			ctx = context.WithValue(ctx, keyNewUser, true)
// 		}
// 		next.ServeHTTP(response, request.WithContext(ctx))
// 	})
// }

// // CheckDBHandler проверяет наличие базы данных.
// func (app *Application) CheckDBHandler(responce http.ResponseWriter, request *http.Request) {
// 	if request.Method == http.MethodGet && app.Storage.PingDB() {
// 		responce.WriteHeader(http.StatusOK)
// 	}
// 	responce.WriteHeader(http.StatusInternalServerError)
// }

// // ShortenHandler сокращает URL полученные в теле запроса POST, принимает text/plain, проверят Content-type, присваивает правильный Content-type ответу,
// // записывает правильный статус в ответ, получает тело запроса и если оно не пустое, то запрашивает сокращенную ссылку
// // и возвращает ответ. Во всех иных случаях возвращает в ответе Bad request.
// func (app *Application) ShortenHandler(responce http.ResponseWriter, request *http.Request) {
// 	logger.Log.Debug("Handler ShortenHandler")
// 	var err error
// 	var body string
// 	// проверяем на content-type
// 	if strings.Contains(request.Header.Get("Content-Type"), "text/plain") || strings.Contains(request.Header.Get("Content-type"), "application/x-gzip") {
// 		// если прошли то присваиваем значение content-type: "text/plain" и статус 201
// 		responce.Header().Add("Content-Type", "text/plain")
// 		// получаем тело запроса
// 		var url []byte
// 		url, err = io.ReadAll(request.Body)
// 		if err != nil {
// 			log.Fatal(err)
// 			responce.WriteHeader(http.StatusBadRequest)
// 			return
// 		}
// 		logger.Log.Debug("Body", zap.String("request URL", string(url)))
// 		// если тело запроса не пустое, то создаем сокращенный url и выводим в тело ответа
// 		if len(url) > 0 {
// 			body, err = app.Storage.CreateShortURL(string(url), app.Config.GetConfig().OuterAddress, request.Context().Value(keyUserID).(int))
// 			if err != nil {
// 				responce.WriteHeader(http.StatusConflict)
// 			}
// 			logger.Log.Debug("Result body", zap.String("sort URL", string(body)))
// 		}
// 		responce.WriteHeader(http.StatusCreated)
// 		responce.Write([]byte(body))
// 		return
// 	}
// 	responce.WriteHeader(http.StatusBadRequest)
// }

// // ShortenJSONHandler сокращает URL полученный в JSON, принимает application/json, проверят Content-type, присваивает правильный Content-type ответу,
// // записывает правильный статус в ответ, получает тело запроса и если оно не пустое, то запрашивает сокращенную ссылку
// // и возвращает ответ. Во всех иных случаях возвращает в ответе Bad request.
// func (app *Application) ShortenJSONHandler(responce http.ResponseWriter, request *http.Request) {
// 	logger.Log.Debug("Handler ShortenJSONHandler")
// 	var err error
// 	var body []byte
// 	// проверяем на content-type
// 	if strings.Contains(request.Header.Get("Content-Type"), "application/json") || strings.Contains(request.Header.Get("Content-type"), "application/x-gzip") {
// 		// если прошли то присваиваем значение content-type: "application/json" и статус 201
// 		responce.Header().Add("Content-Type", "application/json")
// 		// получаем тело запроса
// 		var js []byte
// 		js, err = io.ReadAll(request.Body)
// 		if err != nil {
// 			logger.Log.Error("Request wihtout body", zap.Error(err))
// 			responce.WriteHeader(http.StatusBadRequest)
// 			return
// 		}
// 		logger.Log.Debug("Body", zap.String("type json", string(js)))
// 		// если тело запроса не пустое, то создаем сокращенный url и выводим в тело ответа
// 		if len(js) > 0 {
// 			var url JsRequest
// 			if err = json.Unmarshal(js, &url); err != nil {
// 				logger.Log.Error("Error json parsing", zap.String("request body", string(js)))
// 			}
// 			if url.URL == "" {
// 				responce.WriteHeader(http.StatusCreated)
// 				return
// 			}
// 			extURL, err := app.Storage.CreateShortURL(url.URL, app.Config.GetConfig().OuterAddress, request.Context().Value(keyUserID).(int))
// 			if err != nil {
// 				responce.WriteHeader(http.StatusConflict)
// 			}
// 			result := JsResponce{URL: extURL}
// 			body, err = json.Marshal(result)
// 			if err != nil {
// 				logger.Log.Error("Error json serialization", zap.String("var", fmt.Sprint(result)))
// 			}
// 		}
// 		responce.WriteHeader(http.StatusCreated)
// 		responce.Write(body)
// 		return
// 	}
// 	responce.WriteHeader(http.StatusBadRequest)
// }

// // ShortenBatchJSONHandler сокращает URL полученный в JSON, принимает application/json, проверят Content-type, присваивает правильный Content-type ответу,
// // записывает правильный статус в ответ, получает тело запроса и если оно не пустое, то запрашивает сокращенную ссылку
// // и возвращает ответ. Во всех иных случаях возвращает в ответе Bad request.
// func (app *Application) ShortenBatchJSONHandler(responce http.ResponseWriter, request *http.Request) {
// 	logger.Log.Debug("Handler ShortenBatchJSONHandler")
// 	var err error
// 	var body []byte
// 	// проверяем на content-type
// 	if strings.Contains(request.Header.Get("Content-Type"), "application/json") || strings.Contains(request.Header.Get("Content-type"), "application/x-gzip") {
// 		// если прошли то присваиваем значение content-type: "application/json" и статус 201
// 		responce.Header().Add("Content-Type", "application/json")
// 		// получаем тело запроса
// 		var js []byte
// 		js, err = io.ReadAll(request.Body)
// 		if err != nil {
// 			logger.Log.Error("Request wihtout body", zap.Error(err))
// 			responce.WriteHeader(http.StatusBadRequest)
// 			return
// 		}
// 		logger.Log.Debug("Body", zap.String("type json", string(js)))
// 		// если тело запроса не пустое, то создаем сокращенный url и выводим в тело ответа
// 		// jjs := []byte(`[{"correlation_id": "1","original_url": "ya.ru"}]`)
// 		if len(js) > 0 {
// 			var urls []JsBatchRequest
// 			var resulturls []JsBatchResponce
// 			if err := json.Unmarshal(js, &urls); err != nil {
// 				logger.Log.Error("Error json parsing", zap.String("request body", string(js)))
// 			}
// 			if len(urls) == 0 {
// 				responce.WriteHeader(http.StatusCreated)
// 				return
// 			}
// 			for _, e := range urls {
// 				extURL, _ := app.Storage.CreateShortURL(e.URL, app.Config.GetConfig().OuterAddress, request.Context().Value(keyUserID).(int))
// 				resulturls = append(resulturls, JsBatchResponce{ID: e.ID, SortURL: extURL})
// 			}
// 			body, _ = json.Marshal(resulturls)
// 		}
// 		responce.WriteHeader(http.StatusCreated)
// 		responce.Write(body)
// 		return
// 	}
// 	responce.WriteHeader(http.StatusBadRequest)
// }

// // ExpandHandler получет адрес по короткой ссылке из GET запроса.
// func (app *Application) ExpandHandler(responce http.ResponseWriter, request *http.Request) {
// 	logger.Log.Debug("Handler ExpandHandler")
// 	if request.Method == http.MethodGet {
// 		ctx := request.Context()
// 		outURL, err := app.Storage.GetURL("http://"+app.Config.GetConfig().ServerAddress+request.URL.Path, ctx.Value(keyUserID).(int))
// 		if err == storage.ErrLinkDeleted {
// 			logger.Log.Debug("URL has been deleted", zap.Error(err))
// 			responce.WriteHeader(http.StatusGone)
// 			return
// 		} else if err != nil {
// 			logger.Log.Debug("Can't to get URL", zap.Error(err))
// 			responce.WriteHeader(http.StatusBadRequest)
// 		}
// 		logger.Log.Debug("URL expanded " + outURL)
// 		responce.Header().Add("Location", outURL)
// 		responce.WriteHeader(http.StatusTemporaryRedirect)
// 		return
// 	}
// 	responce.WriteHeader(http.StatusBadRequest)
// }

// // ExpandUserURLSHandler получает весь список сокращенных адресов пользователем прошедшим авторизацию.
// func (app *Application) ExpandUserURLSHandler(responce http.ResponseWriter, request *http.Request) {
// 	logger.Log.Debug("Handler ExpandUserURLSHandler")
// 	var urls []JsUserRequest
// 	ctx := request.Context()

// 	if request.Method == http.MethodGet {

// 		if ctx.Value(keyNewUser) == true {
// 			responce.WriteHeader(http.StatusNoContent)
// 			return
// 		}
// 		urlsr, _ := app.Storage.GetUserURLS(ctx.Value(keyUserID).(int))
// 		for _, e := range urlsr {
// 			urls = append(urls, JsUserRequest{ShortURL: e.ShortURL, OriginalURL: e.URL})
// 		}
// 		if len(urls) == 0 {
// 			responce.WriteHeader(http.StatusNoContent)
// 			return
// 		}
// 		body, err := json.Marshal(urls)
// 		if err != nil {
// 			logger.Log.Debug("can't serialized json answer")
// 		}
// 		responce.Header().Add("Content-Type", "application/json")
// 		responce.WriteHeader(http.StatusOK)
// 		responce.Write(body)
// 		logger.Log.Debug("take body to user urls", zap.String("body", string(body)))
// 		return
// 	}
// 	responce.WriteHeader(http.StatusBadRequest)
// }

// // DeleteUserURLSHandler удалет указанные в JSON сокращенные адреса пользователя прошедшего авторизацию.
// func (app *Application) DeleteUserURLSHandler(responce http.ResponseWriter, request *http.Request) {
// 	logger.Log.Debug("Handler DeleteUserURLSHandler")
// 	var err error
// 	if strings.Contains(request.Header.Get("Content-Type"), "application/json") || strings.Contains(request.Header.Get("Content-type"), "application/x-gzip") {
// 		var js []byte
// 		js, err = io.ReadAll(request.Body)
// 		if err != nil {
// 			logger.Log.Error("Request wihtout body", zap.Error(err))
// 			responce.WriteHeader(http.StatusBadRequest)
// 			return
// 		}
// 		logger.Log.Debug("Body", zap.String("type json", string(js)))
// 		if len(js) > 0 {
// 			var ids struct {
// 				UserID   int
// 				LinksIDS []string
// 			}
// 			ctx := request.Context()
// 			ids.UserID = ctx.Value(keyUserID).(int)
// 			if err := json.Unmarshal(js, &ids.LinksIDS); err != nil {
// 				logger.Log.Error("Error json parsing", zap.String("request body", string(js)))
// 			}
// 			app.Storage.DeleteUserURLS(ids)
// 			responce.WriteHeader(http.StatusAccepted)
// 			return
// 		}
// 	}
// 	responce.WriteHeader(http.StatusBadRequest)
// }

// // GetServerStats показывает статистику сервера если X-Real-IP входит в доверенную подсеть.
// func (app *Application) GetServerStats(responce http.ResponseWriter, request *http.Request) {
// 	logger.Log.Debug("Handler GetServerStats")
// 	var err error
// 	// проверяем установлена ли доверенная подсеть, если нет запрещаем доступ
// 	if !app.Config.TrustedSubnet.Use {
// 		responce.WriteHeader(http.StatusForbidden)
// 		return
// 	}
// 	// берем ip пользователя из заголовка X-Real-IP и проверяем на вхождение его в доверенную подсеть
// 	ip := net.ParseIP(request.Header.Get("X-Real-IP"))
// 	if !app.Config.TrustedSubnet.Path.Contains(ip) {
// 		responce.WriteHeader(http.StatusForbidden)
// 		return
// 	}
// 	// запрашиваем статистику в хранилище
// 	users, err := app.Storage.GetSumUsers()
// 	if err != nil {
// 		logger.Log.Error("Error get statistic", zap.Error(err))
// 		responce.WriteHeader(http.StatusInternalServerError)
// 		return
// 	}
// 	links, err := app.Storage.GetSumURLS()
// 	if err != nil {
// 		logger.Log.Error("Error get statistic", zap.Error(err))
// 		responce.WriteHeader(http.StatusInternalServerError)
// 		return
// 	}
// 	// формируем ответ и отдаем пользователю
// 	js := JsStat{
// 		URLS:  links,
// 		Users: users,
// 	}
// 	body, err := json.Marshal(js)
// 	if err != nil {
// 		logger.Log.Error("Error marshal JSON")
// 		responce.WriteHeader(http.StatusInternalServerError)
// 		return
// 	}

// 	responce.WriteHeader(http.StatusOK)
// 	responce.Write(body)
// }
