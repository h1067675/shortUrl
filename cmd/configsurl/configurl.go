package configsurl

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	NetAddressServerShortener NetAddressServer
	NetAddressServerExpand    NetAddressServer
	FileStoragePath           FilePath
	EnvConf                   EnvConfig
}

type NetAddressServer struct {
	Host string
	Port int
}

type FilePath struct {
	Path string
}

type EnvConfig struct {
	ServerShortener string `env:"SERVER_ADDRESS"`
	ServerExpand    string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
}

// NewConfig - функция создания конфига, получает адреса серверов в виде строки при этом если строки не установлены, то устанавливает
// адреса по умолчанию: localhost:8080
func NewConfig(netAddressServerShortener string, netAddressServerExpand string, fileStoragePath string) *Config {
	var r = Config{ // переменная которая будет хранить сетевой адрес сервера (аргумент -a командной строки)
		NetAddressServerShortener: NetAddressServer{
			// переменная которая будет хранить сетевой адрес сервера (аргумент -a командной строки)
			Host: "localhost",
			Port: 8080,
		},
		NetAddressServerExpand: NetAddressServer{
			// переменная которая будет хранить сетевой адрес подставляемый к сокращенным ссылкам (аргумент -b командной строки)
			Host: "localhost",
			Port: 8080,
		},
		EnvConf: EnvConfig{},
	}
	r.NetAddressServerShortener.Set(netAddressServerShortener)
	r.NetAddressServerExpand.Set(netAddressServerExpand)
	r.FileStoragePath.Set(fileStoragePath)
	return &r
}

// checkNetAddress - функция проверяющая на корректность указания пары host:port и в случае ошибки передающей значения по умолчанию
func checkNetAddress(s string, h string, p int) (host string, port int, e error) {
	host, port = h, p
	v := strings.Split(s, "://")
	if len(v) < 1 || len(v) > 2 {
		e = errors.New("incorrect net address")
		return
	}
	if len(v) == 2 {
		s = v[1]
	}
	a := strings.Split(s, ":")
	if len(a) < 1 || len(a) > 2 {
		e = errors.New("incorrect net address")
		return
	}
	host = a[0]
	ip := net.ParseIP(host)
	if ip == nil && host != "localhost" {
		e = errors.New("incorrect net address")
		return
	}
	if a[1] != "" {
		port, e = strconv.Atoi(a[1])
		if e != nil || port < 0 || port > 65535 {
			e = errors.New("incorrect net address")
			return
		}
	}
	return
}

// возвращаем адрес вида host:port
func (n *NetAddressServer) String() string {
	return fmt.Sprint(n.Host + ":" + strconv.Itoa(n.Port))
}

// устанавливаем значения host и port в переменные
func (n *NetAddressServer) Set(s string) (err error) {
	n.Host, n.Port, err = checkNetAddress(s, n.Host, n.Port)
	if err != nil {
		return err
	}
	return nil
}

func (n *FilePath) Set(s string) (err error) {
	n.Path = s
	return nil
}

// возвращаем путь файла
func (n *FilePath) String() string {
	return n.Path
}

// ParseFlags - разбираем атрибуты командной строки
func (c *Config) ParseFlags() {
	flag.Var(&c.NetAddressServerShortener, "a", "Net address shortener service (host:port)")
	flag.Var(&c.NetAddressServerExpand, "b", "Net address expand service (host:port)")
	flag.Var(&c.FileStoragePath, "f", "File storage path")
	flag.Parse()
}

// EnvConfigSet - забираем переменные окружения и если они установлены то указывам в конфиг из значения
func (c *Config) EnvConfigSet() {
	err := env.Parse(&c.EnvConf)
	if err != nil {
		log.Fatal(err)
	}
	if c.EnvConf.ServerShortener != "" {
		c.NetAddressServerShortener.Set(c.EnvConf.ServerShortener)
	}
	if c.EnvConf.ServerExpand != "" {
		c.NetAddressServerExpand.Set(c.EnvConf.ServerExpand)
	}
}

// Set - инициирует процесс установки настроек
func (c *Config) Set() {
	c.ParseFlags()
	c.EnvConfigSet()
}

// GetConfig - передает структцру со строками адреса сервера сокращения и адреса сервера переадресации коротких адресов
func (c *Config) GetConfig() struct {
	ServerAddress   string
	OuterAddress    string
	FileStoragePath string
} {
	return struct {
		ServerAddress   string
		OuterAddress    string
		FileStoragePath string
	}{ServerAddress: c.NetAddressServerShortener.String(), OuterAddress: c.NetAddressServerExpand.String(), FileStoragePath: c.FileStoragePath.Path}
}
