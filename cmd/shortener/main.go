// Сервис сокращает ссылки и выдает короткую ссылку в ответ
//
// работает как с единичными запросами так и с пакетными запросами через json
// доступна авторизация пользователя и управление сохраненными ссылками
//

package main

import (
	// "log"
	// "net/http"
	// _ "net/http/pprof" // подключаем пакет pprof

	"github.com/h1067675/shortUrl/cmd/configsurl"
	"github.com/h1067675/shortUrl/cmd/storage"
	"github.com/h1067675/shortUrl/internal/handlers"
	"github.com/h1067675/shortUrl/internal/logger"
	"github.com/h1067675/shortUrl/internal/router"
)

// Start - загружает настройки и стартует сервер
func main() {
	// go func() {
	// 	log.Println(http.ListenAndServe("localhost:6060", nil))
	// }()
	// Инициализируем логгер
	logger.Initialize("debug")
	// Устанавливаем настройки приложения по умолчанию
	var conf = configsurl.NewConfig("localhost:8080", "localhost:8080", "./storage.json", "host=127.0.0.1 port=5432 dbname=postgres user=postgres password=12345678 connect_timeout=10 sslmode=prefer")
	// Устанавливаем конфигурацию из параметров запуска или из переменных окружения
	conf.Set()
	// Создаем хранилище данных
	var st = storage.NewStorage(conf.DatabaseDSN.String())
	// Если соединение с базой данных не установлено или не получилось создать таблицу, то загружаем ссылки из файла
	if !st.DB.Connected && conf.FileStoragePath.Path != "" {
		st.RestoreFromfile(conf.FileStoragePath.Path)
	}
	// Создаем соединение и помещвем в него переменные хранения и конфигурации
	var application handlers.Application
	var router router.Router
	router.New()
	application.New(st, conf, router)
	// Запускаем сервер
	application.StartServer()
	// time.Sleep(10 * time.Second)

	// // создаём файл журнала профилирования памяти
	// fmem, err := os.Create(`base.pprof`)
	// if err != nil {
	// 	panic(err)
	// }
	// defer fmem.Close()
	// runtime.GC() // получаем статистику по использованию памяти
	// if err := pprof.WriteHeapProfile(fmem); err != nil {
	// 	panic(err)
	// }
}
