// Package configsurl реализует конфигурирование сервера
package configsurl

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
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
		JSONConfigFile            FilePath
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
		JSONConfigFile  string `env:"CONFIG"`
	}

	JSONConfigParse struct {
		ServerShortener string `json:"server_address"`
		ServerExpand    string `json:"base_url"`
		FileStoragePath string `json:"file_storage_path"`
		DatabaseDSN     string `json:"database_dsn"`
		EnableHTTPS     string `json:"enable_https"`
	}
)

// NewConfig создает конфиг и устанавливает настройки конфигурации в следующем приоритете:
//  1. Параметры запуска
//  2. Переменные среды
//  3. Файл конфигурации
//  4. Настройки по умолчанию
func NewConfig(netAddressServerShortener string, netAddressServerExpand string, fileStoragePath string, dbPath string) (*Config, error) {
	var err error
	var r = Config{}

	// Устанавливаем конфигурацию из переменных окружения
	err = r.EnvConfigSet()
	if err != nil {
		logger.Log.Debug("", zap.String("Errors when setting startup parameters and environment variables", err.Error()))
	}

	// Устанавливаем конфигурацию из параметров запуска
	r.ParseFlags()

	// Заполняем параметры конфигурации не получившие значения из переменных среды или параметров запуска
	err1 := r.SetConfigFromFileOrDefault(netAddressServerShortener, netAddressServerExpand, fileStoragePath, dbPath)
	if err1 != nil {
		err = errors.Join(err, err1)
	}
	return &r, err
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
	if n.Port > 0 {
		return fmt.Sprint(n.Host + ":" + strconv.Itoa(n.Port))
	}
	return fmt.Sprint(n.Host)
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

// Set cохраняет значение переменной среды.
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
	flag.Var(&c.JSONConfigFile, "config", "Sets the path to the configuration file in JSON format")
	flag.Var(&c.JSONConfigFile, "c", "reduction to -config flag")
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
	if c.EnvConf.EnableHTTPS != "" {
		c.EnableHTTPS.On = true
	}
	if c.EnvConf.FileStoragePath != "" {
		err1 := c.FileStoragePath.Set(c.EnvConf.FileStoragePath)
		err = errors.Join(err, err1)
	}
	return
}

// SetConfigFromFileOrDefault заполняет параметры конфигурации не получившие значения из переменных среды или параметров запуска
func (c *Config) SetConfigFromFileOrDefault(netAddressServerShortener string, netAddressServerExpand string, fileStoragePath string, dbPath string) error {
	var err error

	// Проверяем есть ли в конфигурации файл с настройками JSON, если есть то читаем из него данные
	var jscfg JSONConfigParse
	if c.JSONConfigFile.String() != "" {
		var err1 error
		jscfg, err1 = c.GetConfigFromJSONFile()
		err = errors.Join(err, err1)
	}

	// перебираем все параметры и если есть параметры без значений заполняем
	// их данными изначально из файла настроек, затем из настроек по умолчанию
	if c.NetAddressServerExpand.String() == "" {
		if jscfg.ServerExpand != "" {
			err = errors.Join(err, c.NetAddressServerExpand.Set(jscfg.ServerExpand))
		} else {
			err = errors.Join(err, c.NetAddressServerExpand.Set(netAddressServerExpand))
		}
	}
	if c.NetAddressServerShortener.String() == "" {
		if jscfg.ServerShortener != "" {
			err = errors.Join(err, c.NetAddressServerShortener.Set(jscfg.ServerShortener))
		} else {
			err = errors.Join(err, c.NetAddressServerShortener.Set(netAddressServerShortener))
		}
	}
	if c.FileStoragePath.String() == "" {
		if jscfg.FileStoragePath != "" {
			err = errors.Join(err, c.FileStoragePath.Set(jscfg.FileStoragePath))
		} else {
			err = errors.Join(err, c.FileStoragePath.Set(fileStoragePath))
		}
	}
	if c.DatabaseDSN.String() == "" {
		if jscfg.DatabaseDSN != "" {
			err = errors.Join(err, c.DatabaseDSN.Set(jscfg.DatabaseDSN))
		} else {
			err = errors.Join(err, c.DatabaseDSN.Set(dbPath))
		}
	}
	if c.EnableHTTPS.String() == "" {
		if jscfg.EnableHTTPS != "" {
			c.EnableHTTPS.On = true
		}
	}
	return err
}

// GetConfigFromJSONFile импортирует настройки из файла конфигурации
func (c *Config) GetConfigFromJSONFile() (JSONConfigParse, error) {
	var cfg JSONConfigParse
	var err error
	// получаем директорию текущего файла
	appfile, err := os.Executable()
	if err != nil {
		return cfg, err
	}
	// читаем файл конфигурации
	data, err := os.ReadFile(filepath.Join(filepath.Dir(appfile), c.JSONConfigFile.String()))
	if err != nil {
		return cfg, err
	}
	// помещаем настройки из файла в структуру
	if err = json.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
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
