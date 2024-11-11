package main

import (
	"github.com/h1067675/shortUrl/cmd/configsurl"
	"github.com/h1067675/shortUrl/cmd/netservice"
	"github.com/h1067675/shortUrl/cmd/storage"
	"github.com/h1067675/shortUrl/internal/logger"
)

func main() {
	// Устанавливаем настройки приложения по умолчанию
	var conf = configsurl.NewConfig("localhost:8080", "localhost:8080", "/storage.json")
	// Устанавливаем конфигурацию из параметров запуска или из переменных окружения
	conf.Set()
	// Создаем хранилище данных
	var storage = storage.NewStorage(conf.DatabaseDSN.String())
	// Если соединение с базой данных не установлено или не получилось создать таблицу, то загружаем ссылки из файла
	if !storage.DB.Connected {
		storage.RestoreFromfile(conf.FileStoragePath.Path)
		defer storage.SaveToFile(conf.GetConfig().FileStoragePath)
	}
	// Создаем соединение и помещвем в него переменные хранения и конфигурации
	var conn = netservice.NewConnect(storage, conf)
	// Инициализируем логгер
	logger.Initialize("debug")
	// Запускаем сервер
	conn.StartServer()
}
