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

	"fmt"

	"github.com/h1067675/shortUrl/cmd/configsurl"
	"github.com/h1067675/shortUrl/cmd/storage"
	"github.com/h1067675/shortUrl/internal/app"
	"github.com/h1067675/shortUrl/internal/logger"
	"github.com/h1067675/shortUrl/internal/router"
	"go.uber.org/zap"
)

// Определяем глобальные переменные для вывода версии сборки указаннной при компиляции
var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	// Получаем глобальные переменные версии и выводим в консоль
	if buildVersion == "" {
		buildVersion = "N/A"
	}
	if buildDate == "" {
		buildDate = "N/A"
	}
	if buildCommit == "" {
		buildCommit = "N/A"
	}
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	// go func() {
	// 	log.Println(http.ListenAndServe("localhost:6060", nil))
	// }()

	// Инициализируем логгер
	err := logger.Initialize("debug")
	if err != nil {
		fmt.Print(err)
	}

	// Устанавливаем настройки приложения по умолчанию
	conf, err := configsurl.NewConfig("localhost:8080", "localhost:8080", "./storage.js", "host=127.0.0.1 port=5432 dbname=postgres user=postgres password=12345678 connect_timeout=10 sslmode=disable")
	if err != nil {
		logger.Log.Debug("Errors when configuring the server", zap.String("Error", err.Error()))
	}

	// Создаем хранилище данных
	var st = storage.NewStorage(conf.DatabaseDSN.String())

	// Если соединение с базой данных не установлено или не получилось создать таблицу, то загружаем ссылки из файла
	if !st.DB.Connected && conf.FileStoragePath.Path != "" {
		st.RestoreFromfile(conf.FileStoragePath.Path)
	}

	// Создаем соединение и маршрутизацию
	serverHTTP := router.New()

	// запускаем бизнес-логику и помещвем в нее переменные хранения, конфигурации и маршрутизации
	var application app.Application
	application.New(st, conf, serverHTTP)

	// Запускаем сервер
	application.StartServers()

	// time.Sleep(10 * time.Second)
	// // создаём файл журнала профилирования памяти
	// fmem, err := os.Create(`profiles/base.pprof`)
	// if err != nil {
	// 	panic(err)
	// }
	// defer fmem.Close()
	// runtime.GC() // получаем статистику по использованию памяти
	// if err := pprof.WriteHeapProfile(fmem); err != nil {
	// 	panic(err)
	// }

}
