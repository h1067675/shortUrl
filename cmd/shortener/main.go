// Сервис сокращает ссылки и выдает короткую ссылку в ответ
//
// работает как с единичными запросами так и с пакетными запросами через json
// доступна авторизация пользователя и управление сохраненными ссылками
//

package main

import (
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/h1067675/shortUrl/cmd/configsurl"
	"github.com/h1067675/shortUrl/cmd/storage"
	"github.com/h1067675/shortUrl/internal/logger"
)

// Start - загружает настройки и стартует сервер
func main() {
	// Инициализируем логгер
	logger.Initialize("debug")
	// Устанавливаем настройки приложения по умолчанию
	var conf = configsurl.NewConfig("localhost:8080", "localhost:8080", "./storage.json", "host=127.0.0.1 port=5432 dbname=postgres user=postgres password=12345678 connect_timeout=10 sslmode=prefer")
	// Устанавливаем конфигурацию из параметров запуска или из переменных окружения
	conf.Set()
	// Создаем хранилище данных
	var storage = storage.NewStorage(conf.DatabaseDSN.String())
	// Если соединение с базой данных не установлено или не получилось создать таблицу, то загружаем ссылки из файла
	if !storage.DB.Connected && conf.FileStoragePath.Path != "" {
		storage.RestoreFromfile(conf.FileStoragePath.Path)
	}
	// Создаем соединение и помещвем в него переменные хранения и конфигурации
	var conn = NewConnect(storage, conf)
	// Запускаем сервер
	go conn.StartServer()
	time.Sleep(30 * time.Second)
	// создаём файл журнала профилирования памяти
	fmem, err := os.Create(`./profiles/result.pprof`)
	if err != nil {
		panic(err)
	}
	defer fmem.Close()
	runtime.GC() // получаем статистику по использованию памяти
	if err := pprof.WriteHeapProfile(fmem); err != nil {
		panic(err)
	}
}
