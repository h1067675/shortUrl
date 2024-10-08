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
	EnvConf                   EnvConfig
}

type NetAddressServer struct {
	Host string
	Port int
}

type EnvConfig struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseURL       string `env:"BASE_URL"`
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
	if ip == nil {
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

// ParseFlags - разбираем атрибуты командной строки
func (c *Config) ParseFlags() {
	flag.Var(&c.NetAddressServerShortener, "a", "Net address shortener service (host:port)")
	flag.Var(&c.NetAddressServerExpand, "b", "Net address expand service (host:port)")
	flag.Parse()
}

// EnvConfigSet - забираем переменные окружения и если они установлены то указывам в конфиг из значения
func (e *Config) EnvConfigSet() {
	err := env.Parse(&e.EnvConf)
	if err != nil {
		log.Fatal(err)
	}
	if e.EnvConf.ServerAddress != "" {
		e.NetAddressServerShortener.Set(e.EnvConf.ServerAddress)
	}
	if e.EnvConf.BaseURL != "" {
		e.NetAddressServerExpand.Set(e.EnvConf.BaseURL)
	}
}
