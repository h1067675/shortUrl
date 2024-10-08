package main

import (
	"log"
	"net/http"

	"github.com/h1067675/shortUrl/cmd/configsurl"
	"github.com/h1067675/shortUrl/cmd/netservice"
	"github.com/h1067675/shortUrl/cmd/storage"
)

// Создаем соединение и помещвем в него переменные хранения и конфигурации
var conn = netservice.Connect{
	Base: storage.Storage{
		InnerLinks:  make(map[string]string),
		OutterLinks: make(map[string]string),
	},
	Conf: configsurl.Config{ // переменная которая будет хранить сетевой адрес сервера (аргумент -a командной строки)
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
	},
}

func main() {
	conn.Conf.ParseFlags()
	conn.Conf.EnvConfigSet()
	log.Fatal(http.ListenAndServe(conn.Conf.NetAddressServerShortener.String(), conn.RouterFunc()))
}
