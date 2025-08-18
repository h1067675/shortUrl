package router

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/h1067675/shortUrl/internal/compress"
	"github.com/h1067675/shortUrl/internal/logger"
)

// структуры
type (
	// Handler связывает с хэндлерами
	Handler interface {
		Authorization(next http.Handler) http.Handler
		CheckDBHandler(responce http.ResponseWriter, request *http.Request)
		ShortenHandler(responce http.ResponseWriter, request *http.Request)
		ShortenJSONHandler(responce http.ResponseWriter, request *http.Request)
		ShortenBatchJSONHandler(responce http.ResponseWriter, request *http.Request)
		ExpandHandler(responce http.ResponseWriter, request *http.Request)
		ExpandUserURLSHandler(responce http.ResponseWriter, request *http.Request)
		DeleteUserURLSHandler(responce http.ResponseWriter, request *http.Request)
	}
	// Router отвечает за маршрутизацию
	Router struct {
		chi.Router
	}
)

// // NewConnect создает коннектор
// func (r *Router) New() {
// 	r = &Router{
// 		chi.NewRouter(),
// 	}
// }

// RouterFunc делает маршрутизацию к хандлерам.
func (r Router) RouterFunc(handlers Handler) chi.Router {
	// Создаем chi роутер
	r.Router = chi.NewRouter()
	// Добавляем все функции middleware
	r.Router.Use(handlers.Authorization)
	r.Router.Use(compress.CompressHandle)
	r.Router.Use(logger.RequestLogger)

	// Делаем маршрутизацию
	r.Router.Route("/", func(r chi.Router) {
		r.Post("/", handlers.ShortenHandler) // POST запрос отправляем на сокращение ссылки
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", handlers.ExpandHandler) // GET запрос с id направляем на извлечение ссылки
		})
		r.Route("/api/shorten", func(r chi.Router) {
			r.Post("/", handlers.ShortenJSONHandler)           // POST запрос с JSON телом
			r.Post("/batch", handlers.ShortenBatchJSONHandler) // POST запрос с множественным JSON телом
		})
		r.Route("/api/user", func(r chi.Router) {
			r.Get("/urls", handlers.ExpandUserURLSHandler)    // GET запрос на выдачу всех сокращенных ссылок пользователем
			r.Delete("/urls", handlers.DeleteUserURLSHandler) // DELETE запрос удаляет ссылки перечисленные в запросе
		})
		r.Route("/ping", func(r chi.Router) {
			r.Get("/", handlers.CheckDBHandler) // GET запрос проверяет работоспособность базы данных
		})
	})
	return r.Router
}
