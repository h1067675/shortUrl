package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/h1067675/shortUrl/cmd/configsurl"
	"github.com/h1067675/shortUrl/cmd/netservice"
	"github.com/h1067675/shortUrl/cmd/storage"
)

// Создаем хранилище данных
var base = storage.Storage{
	InnerLinks:  make(map[string]string),
	OutterLinks: make(map[string]string),
}

// Устанавливаем настройки приложения по умолчанию
var conf = configsurl.Config{ // переменная которая будет хранить сетевой адрес сервера (аргумент -a командной строки)
	NetAddressServerShortener: configsurl.NetAddressServer{
		// переменная которая будет хранить сетевой адрес сервера (аргумент -a командной строки)
		Host: "localhost",
		Port: 8080,
	},
	NetAddressServerExpand: configsurl.NetAddressServer{
		// переменная которая будет хранить сетевой адрес подставляемый к сокращенным ссылкам (аргумент -b командной строки)
		Host: "localhost",
		Port: 8080,
	},
	EnvConf: configsurl.EnvConfig{},
}

// Создаем соединение и помещвем в него переменные хранения и конфигурации
var conn = netservice.Connect{
	Base: &base,
	Conf: &conf,
}

func main() {
	conf.ParseFlags()
	fmt.Print(conf)
	conf.EnvConfigSet()
	fmt.Println(conf)
	log.Fatal(http.ListenAndServe(conf.NetAddressServerShortener.String(), conn.RouterFunc()))
}
