package app

import (
	"github.com/h1067675/shortUrl/cmd/configsurl"
	"github.com/h1067675/shortUrl/cmd/storage"
	"github.com/h1067675/shortUrl/internal/router"
)

// структуры
type (
	// Application описывает структуру зависимостей для доступа к базе данных и настройкам приложения
	Application struct {
		Storage    *storage.Storage
		Config     *configsurl.Config
		HTTPServer router.Router
	}
)
