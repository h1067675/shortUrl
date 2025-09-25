package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/h1067675/shortUrl/cmd/authorization"
	"github.com/h1067675/shortUrl/cmd/configsurl"
	"github.com/h1067675/shortUrl/cmd/storage"
	"github.com/h1067675/shortUrl/internal/grpcserver"
	"github.com/h1067675/shortUrl/internal/logger"
	"github.com/h1067675/shortUrl/internal/router"

	"go.uber.org/zap"
)

// структуры
type (
	// Application описывает структуру зависимостей для доступа к базе данных, конфигурации и серверам
	Application struct {
		Storage    *storage.Storage
		Config     *configsurl.Config
		ServerHTTP router.Router
		ServerGRPC grpcserver.GrpcServer
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

	// JsStat определяет фотмат выдачи статистики сервера
	JsStat struct {
		Users int `json:"users"`
		URLS  int `json:"urls"`
	}
)

// New задает зависимости между пакетами
func (app *Application) New(s *storage.Storage, c *configsurl.Config, r router.Router) { //, g grpcserver.GrpcServer) {
	app.Storage = s
	app.Config = c
	app.ServerHTTP = r
	// app.ServerGRPC = g
}

// StartServer запускает серверы.
func (app *Application) StartServers() {
	// Определяем HTTP сервер и указываем адрес и ручку
	server := app.ServerHTTP.NewServerHTTP(*app.Config, app)

	// Запускаем ожидание сигналов на завершение работы
	idleConnsClosed := app.waitSysSignals(server)

	// Запускаем HTTP сервер
	go app.ServerHTTP.StartServerHTTP(*app.Config)

	// Ждем мягкого завершения работы сервера
	<-idleConnsClosed

	// Сохраняем хранилище из мапы в файл
	app.Storage.SaveToFile(app.Config.GetConfig().FileStoragePath)
	logger.Log.Info("Storage saved to " + app.Config.GetConfig().FileStoragePath)

	// Сообщаем об окончании работы сервера
	logger.Log.Info("Server has shutdown in graceful mode")
}

// WaitSysSignals определяет логику для отслеживания сигналов завершения работы приложения и
// реализует процесс мягкого завершения работы приложения.
func (app *Application) waitSysSignals(server *http.Server) chan struct{} {
	// через этот канал сообщим основному потоку, что соединения закрыты
	idleConnsClosed := make(chan struct{})
	// канал для перенаправления прерываний
	sigint := make(chan os.Signal, 1)
	// регистрируем перенаправление прерываний
	signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	// запускаем горутину обработки пойманных прерываний
	go func() {
		// читаем из канала прерываний
		<-sigint

		// создаем контекст с таймаутом на завершение операций сервером
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()

		// получили сигнал, запускаем процедуру graceful shutdown сервера
		if err := server.Shutdown(ctx); err != nil {
			logger.Log.Debug(err.Error(), zap.String("server address", app.Config.GetConfig().ServerAddress))
		}

		// сообщаем основному потоку, что все сетевые соединения обработаны и закрыты
		close(idleConnsClosed)
	}()
	return idleConnsClosed
}

// Authorization mildware осуществляет авторизацию пользователя.
func (app *Application) Authorization(cookie string, hasToken bool) (token string, userid int, err error) {
	logger.Log.Debug("checking authorization")
	if hasToken {
		userid, err = authorization.CheckToken(cookie)
		if err != nil {
			return "", 0, err
		}
	} else {
		userid, err := app.Storage.GetNewUserID()
		logger.Log.Debug("new user", zap.Int("id", userid))
		if err != nil {
			logger.Log.Error("don't can to get new user ID", zap.Error(err))
			return "", 0, err
		}
		token, err := authorization.SetToken(userid)
		if err != nil {
			logger.Log.Error("don't can to create token", zap.Error(err))
			return "", 0, err
		}
		return token, userid, nil
	}
	return "", userid, nil
}

// CheckDB проверяет наличие базы данных.
func (app *Application) CheckDB() (statuscode int) {
	if app.Storage.PingDB() {
		return http.StatusOK
	}
	return http.StatusInternalServerError
}

// Shorten сокращает URL полученные в теле запроса POST, принимает text/plain, проверят Content-type, присваивает правильный Content-type ответу,
// записывает правильный статус в ответ, получает тело запроса и если оно не пустое, то запрашивает сокращенную ссылку
// и возвращает ответ. Во всех иных случаях возвращает в ответе Bad request.
func (app *Application) Shorten(url string, userid int) (body []byte, statusCode int) {
	if len(url) > 0 {
		body, err := app.Storage.CreateShortURL(string(url), app.Config.GetConfig().OuterAddress, userid)
		if err != nil {
			return nil, http.StatusConflict
		}
		logger.Log.Debug("Result body", zap.String("sort URL", string(body)))
	}
	return body, http.StatusCreated
}

// ShortenJSONHandler сокращает URL полученный в JSON, принимает application/json, проверят Content-type, присваивает правильный Content-type ответу,
// записывает правильный статус в ответ, получает тело запроса и если оно не пустое, то запрашивает сокращенную ссылку
// и возвращает ответ. Во всех иных случаях возвращает в ответе Bad request.
func (app *Application) ShortenJSON(js []byte, userid int) (body []byte, statusCode int) {
	var err error
	if len(js) > 0 {
		var url JsRequest
		if err = json.Unmarshal(js, &url); err != nil {
			logger.Log.Error("Error json parsing", zap.String("request body", string(js)))
			return nil, http.StatusInternalServerError
		}
		if url.URL == "" {
			return nil, http.StatusCreated
		}
		extURL, err := app.Storage.CreateShortURL(url.URL, app.Config.GetConfig().OuterAddress, userid)
		if err != nil {
			return nil, http.StatusConflict
		}
		result := JsResponce{URL: extURL}
		body, err = json.Marshal(result)
		if err != nil {
			logger.Log.Error("Error json serialization", zap.String("var", fmt.Sprint(result)))
			return nil, http.StatusInternalServerError
		}
		return body, http.StatusCreated
	}
	return nil, http.StatusBadRequest
}

// ShortenBatchJSONHandler сокращает URL полученный в JSON, принимает application/json, проверят Content-type, присваивает правильный Content-type ответу,
// записывает правильный статус в ответ, получает тело запроса и если оно не пустое, то запрашивает сокращенную ссылку
// и возвращает ответ. Во всех иных случаях возвращает в ответе Bad request.
func (app *Application) ShortenBatchJSON(js []byte, userid int) (body []byte, statusCode int) {
	if len(js) > 0 {
		var urls []JsBatchRequest
		var resulturls []JsBatchResponce
		if err := json.Unmarshal(js, &urls); err != nil {
			logger.Log.Error("Error json parsing", zap.String("request body", string(js)))
			return nil, http.StatusInternalServerError
		}
		if len(urls) == 0 {
			return nil, http.StatusCreated
		}
		for _, e := range urls {
			extURL, _ := app.Storage.CreateShortURL(e.URL, app.Config.GetConfig().OuterAddress, userid)
			resulturls = append(resulturls, JsBatchResponce{ID: e.ID, SortURL: extURL})
		}
		body, err := json.Marshal(resulturls)
		if err != nil {
			logger.Log.Error("Error json serialization")
			return nil, http.StatusInternalServerError
		}
		return body, http.StatusCreated
	}
	return nil, http.StatusBadRequest
}

// ExpandHandler получет адрес по короткой ссылке из GET запроса.
func (app *Application) Expand(shortCode string, userid int) (basedURL string, statusCode int) {
	outURL, err := app.Storage.GetURL("http://"+app.Config.GetConfig().ServerAddress+shortCode, userid)
	if err == storage.ErrLinkDeleted {
		logger.Log.Debug("URL has been deleted", zap.Error(err))
		return "", http.StatusGone
	} else if err != nil {
		logger.Log.Debug("Can't to get URL", zap.Error(err))
		return "", http.StatusGone
	}
	logger.Log.Debug("URL expanded " + outURL)
	return "", http.StatusTemporaryRedirect
}

// ExpandUserURLSHandler получает весь список сокращенных адресов пользователем прошедшим авторизацию.
func (app *Application) ExpandUserURLS(userid int, newuser bool) (body []byte, statusCode int) {
	if newuser {
		return nil, http.StatusNoContent
	}
	urlsr, err := app.Storage.GetUserURLS(userid)
	if err != nil {
		return nil, http.StatusInternalServerError
	}
	var urls []JsUserRequest
	for _, e := range urlsr {
		urls = append(urls, JsUserRequest{ShortURL: e.ShortURL, OriginalURL: e.URL})
	}
	if len(urls) == 0 {
		return nil, http.StatusNoContent
	}
	body, err = json.Marshal(urls)
	if err != nil {
		logger.Log.Debug("can't serialized json answer")
		return nil, http.StatusInternalServerError
	}
	return body, http.StatusOK

}

// DeleteUserURLSHandler удалет указанные в JSON сокращенные адреса пользователя прошедшего авторизацию.
func (app *Application) DeleteUserURLS(js []byte, userid int) (statusCode int) {
	if len(js) > 0 {
		var ids struct {
			UserID   int
			LinksIDS []string
		}
		ids.UserID = userid
		if err := json.Unmarshal(js, &ids.LinksIDS); err != nil {
			logger.Log.Error("Error json parsing", zap.String("request body", string(js)))
			return http.StatusInternalServerError
		}
		err := app.Storage.DeleteUserURLS(ids)
		if err != nil {
			logger.Log.Error("Error delete URLS", zap.String("request body", string(js)))
			return http.StatusInternalServerError
		}
		return http.StatusAccepted
	}
	return http.StatusBadRequest
}

// GetServerStats показывает статистику сервера если X-Real-IP входит в доверенную подсеть.
func (app *Application) GetServerStats(ip net.IP) (body []byte, statusCode int) {
	var err error

	// проверяем установлена ли доверенная подсеть, если нет запрещаем доступ
	if !app.Config.TrustedSubnet.Use {
		return nil, http.StatusForbidden
	}

	// Сверяем ip пользователя на вхождение его в доверенную подсеть
	if !app.Config.TrustedSubnet.Path.Contains(ip) {
		return nil, http.StatusForbidden
	}

	// запрашиваем статистику в хранилище
	users, err := app.Storage.GetSumUsers()
	if err != nil {
		logger.Log.Error("Error get statistic", zap.Error(err))
		return nil, http.StatusInternalServerError
	}
	links, err := app.Storage.GetSumURLS()
	if err != nil {
		logger.Log.Error("Error get statistic", zap.Error(err))
		return nil, http.StatusInternalServerError
	}

	// формируем ответ и отдаем пользователю
	js := JsStat{
		URLS:  links,
		Users: users,
	}
	body, err = json.Marshal(js)
	if err != nil {
		logger.Log.Error("Error marshal JSON")
		return nil, http.StatusInternalServerError
	}

	return body, http.StatusOK
}
