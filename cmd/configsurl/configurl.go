// Package configsurl реализует конфигурирование сервера
package configsurl

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/caarlos0/env/v6"
	"github.com/h1067675/shortUrl/internal/logger"
	"go.uber.org/zap"
)

// структуры
type (
	// Config определяет настройки сервера.
	Config struct {
		NetAddressServerShortener NetAddressServer
		NetAddressServerExpand    NetAddressServer
		FileStoragePath           FilePath
		DatabaseDSN               DatabasePath
		EnableHTTPS               EnableHTTPS
		EnvConf                   EnvConfig
	}

	// NetAddressServer описывает формат сетевого адреса для получения переменной среды.
	NetAddressServer struct {
		Host string
		Port int
	}

	// FilePath описывает формат пути к файлу сохранения для получения переменной среды.
	FilePath struct {
		Path string
	}

	// DatabasePath описывает формат адреса подключения к БД.
	DatabasePath struct {
		Path string
	}

	// EnableHTTPS определяет настройку использования HTTPS
	EnableHTTPS struct {
		On bool
	}

	// EnvConfig описывает название переменных среды.
	EnvConfig struct {
		ServerShortener string `env:"SERVER_ADDRESS"`
		ServerExpand    string `env:"BASE_URL"`
		FileStoragePath string `env:"FILE_STORAGE_PATH"`
		DatabaseDSN     string `env:"DATABASE_DSN"`
		EnableHTTPS     string `env:"ENABLE_HTTPS"`
	}
)

// NewConfig создает конфиг и получает адреса серверов в виде строки при этом если строки не установлены, то устанавливает
// адреса по умолчанию: localhost:8080.
func NewConfig(netAddressServerShortener string, netAddressServerExpand string, fileStoragePath string, dbPath string) (*Config, error) {
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
	err1 := r.NetAddressServerShortener.Set(netAddressServerShortener)
	err2 := r.NetAddressServerExpand.Set(netAddressServerExpand)
	err3 := r.FileStoragePath.Set(fileStoragePath)
	err4 := r.DatabaseDSN.Set(dbPath)
	return &r, errors.Join(err1, err2, err3, err4)
}

// checkNetAddress проверяtn на корректность указания пары host:port и в случае ошибки передающей значения по умолчанию.
func checkNetAddress(s string) (host string, port int, e error) {
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

// String возвращает адрес вида host:port.
func (n *NetAddressServer) String() string {
	return fmt.Sprint(n.Host + ":" + strconv.Itoa(n.Port))
}

// Set устанавливет значения host и port в переменные.
func (n *NetAddressServer) Set(s string) (err error) {
	p, h, err := checkNetAddress(s)
	n.Host, n.Port = p, h
	return err
}

// Set сохраняет значение переменной среды.
func (n *FilePath) Set(s string) (err error) {
	n.Path = s
	return nil
}

// String возвращает путь файла.
func (n *FilePath) String() string {
	return n.Path
}

// Set cохраняет значение переменной среды,
func (n *DatabasePath) Set(s string) (err error) {
	n.Path = s
	return nil
}

// String возвращает путь к базе данных.
func (n *DatabasePath) String() string {
	return n.Path
}

// String возвращает путь файла.
func (n *EnableHTTPS) String() string {
	if n.On {
		return "HTTPS enabled"
	}
	return "HTTPS disabled"
}

// ParseFlags разбирает атрибуты командной строки.
func (c *Config) ParseFlags() {
	flag.Var(&c.NetAddressServerShortener, "a", "Net address shortener service (host:port)")
	flag.Var(&c.NetAddressServerExpand, "b", "Net address expand service (host:port)")
	flag.Var(&c.FileStoragePath, "f", "File storage path")
	flag.Var(&c.DatabaseDSN, "d", "Database path")
	flag.BoolVar(&c.EnableHTTPS.On, "e", false, "Enable HTTPS")
	flag.Parse()
}

// EnvConfigSet забирает переменные окружения и если они установлены и сохраняет в конфиг.
func (c *Config) EnvConfigSet() (err error) {
	err = env.Parse(&c.EnvConf)
	if err != nil {
		logger.Log.Debug("Error parse ENVs", zap.String("Error", err.Error()))
	}
	if c.EnvConf.ServerShortener != "" {
		err1 := c.NetAddressServerShortener.Set(c.EnvConf.ServerShortener)
		err = errors.Join(err, err1)
	}
	if c.EnvConf.ServerExpand != "" {
		err1 := c.NetAddressServerExpand.Set(c.EnvConf.ServerExpand)
		err = errors.Join(err, err1)
	}
	if c.EnvConf.FileStoragePath != "" {
		err1 := c.FileStoragePath.Set(c.EnvConf.FileStoragePath)
		err = errors.Join(err, err1)
	}
	if c.EnvConf.DatabaseDSN != "" {
		err1 := c.DatabaseDSN.Set(c.EnvConf.DatabaseDSN)
		err = errors.Join(err, err1)
	}
	return
}

// Set инициирует процесс установки настроек.
func (c *Config) Set() error {
	c.ParseFlags()
	err := c.EnvConfigSet()
	return err
}

// GetConfig возвращает данные настроек в текстовом формате.
func (c *Config) GetConfig() struct {
	ServerAddress   string
	OuterAddress    string
	FileStoragePath string
	DatabasePath    string
} {
	return struct {
		ServerAddress   string
		OuterAddress    string
		FileStoragePath string
		DatabasePath    string
	}{ServerAddress: c.NetAddressServerShortener.String(),
		OuterAddress:    c.NetAddressServerExpand.String(),
		FileStoragePath: c.FileStoragePath.Path,
		DatabasePath:    c.DatabaseDSN.Path}
}
