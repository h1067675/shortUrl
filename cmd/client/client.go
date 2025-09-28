// Package client реализует клиент для доступа к API
package client

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/h1067675/shortUrl/cmd/configsurl"
	"github.com/h1067675/shortUrl/cmd/storage"
	"github.com/h1067675/shortUrl/internal/handlers"
	"github.com/h1067675/shortUrl/internal/logger"
	"github.com/h1067675/shortUrl/internal/router"
	"go.uber.org/zap"
)

// типы двнных запроса
const (
	JSONData = "application/json"
	TextData = "text/plain"
)

// Client описывает структуру настроек сервера
type Client struct {
	NetAddressServer string
	NetAddressExpand string
	PathStorageFile  string
	PathDateBase     string
	LevelLogging     string
}

// NewClient создает нового клиента
func NewClient() *Client {
	return &Client{}
}

// SetNetAddressServer устанавливает адрес сервера для сокращения ссылок
func (c *Client) SetNetAddressServer(netAddressServer string) *Client {
	c.NetAddressServer = netAddressServer
	return c
}

// SetNetAddressExpans устанавливает адрес сервера извлечения исходного адреса из короткой ссылки
func (c *Client) SetNetAddressExpans(netAddressExpand string) *Client {
	c.NetAddressExpand = netAddressExpand
	return c
}

// SetPathStorageFile устанавливает путь для хранения резервного файла данных
func (c *Client) SetPathStorageFile(pathStorageFile string) *Client {
	c.PathStorageFile = pathStorageFile
	return c
}

// SetPathDateBase устанавливает путь доступа к базе данных
func (c *Client) SetPathDateBase(pathDateBase string) *Client {
	c.PathDateBase = pathDateBase
	return c
}

// SetLevelLogging устанавливает уровень сбора логов
func (c *Client) SetLevelLogging(levelLogging string) *Client {
	c.LevelLogging = levelLogging
	return c
}

// StartServer запускает сервер сокращения ссылок
func (c Client) StartServer() {
	if c.NetAddressServer == "" {
		c.NetAddressServer = "localhost:8080"
	}
	if c.NetAddressExpand == "" {
		c.NetAddressExpand = "localhost:8080"
	}
	if c.PathStorageFile == "" {
		c.PathStorageFile = "./storage.json"
	}
	if c.PathDateBase == "" {
		c.PathDateBase = "host=127.0.0.1 port=5432 dbname=postgres user=postgres password=12345678 connect_timeout=10 sslmode=prefer"
	}
	if c.LevelLogging == "" {
		c.LevelLogging = "info"
	}
	// Инициализируем логгер
	err := logger.Initialize(c.LevelLogging)
	if err != nil {
		fmt.Print(err)
	}
	// Устанавливаем настройки приложения по умолчанию
	conf, err := configsurl.NewConfig(c.NetAddressServer, c.NetAddressExpand, c.PathStorageFile, c.PathDateBase)
	if err != nil {
		logger.Log.Debug("Уrrors when configuring the server", zap.String("Error", err.Error()))
	}

	// Создаем хранилище данных
	var st = storage.NewStorage(conf.DatabaseDSN.String())
	// Если соединение с базой данных не установлено или не получилось создать таблицу, то загружаем ссылки из файла
	if !st.DB.Connected && conf.FileStoragePath.Path != "" {
		st.RestoreFromfile(conf.FileStoragePath.Path)
	}
	// Создаем соединение и маршрутизацию
	var router router.Router
	// запускаем бизнес-логику и помещвем в нее переменные хранения, конфигурации и маршрутизации
	var application handlers.Application
	application.New(st, conf, router)
	// Запускаем сервер
	go application.StartServer()
}

// Request делает сетевой запрос методом method, по адресу endpoint,
func (c *Client) request(method string, endpoint string, cType string, body string) (string, error) {
	// контейнер данных для запроса
	data := url.Values{}
	// заполняем контейнер данными если запрос будет совершен методом POST
	if method == "POST" {
		data.Set("url", body)
	}
	// добавляем HTTP-клиент
	client := &http.Client{}
	// пишем запрос
	request, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	// в заголовках запроса указываем кодировку
	request.Header.Add("Content-Type", cType)
	// отправляем запрос и получаем ответ
	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	// выводим код ответа
	fmt.Println("Статус-код ", response.Status)
	defer func() {
		err1 := response.Body.Close()
		if err1 != nil {
			logger.Log.Debug("Errors to close Body", zap.String("Error", err1.Error()))
		}
	}()
	// читаем поток из тела ответа
	result, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	// и печатаем его
	return string(result), nil
}

// Post делает POST-запрос к адресной строке endpoint с телом запроса body и типом контента cType.
// Виды типов контента:
// jsonData = "application/json"
// textData = "text/plain"
func (c Client) Post(endpoint string, body string, cType string) (string, error) {
	return c.request(http.MethodPost, endpoint, cType, body)
}

// Get делает GET-запрос к адресной строке endpoint и типом контента cType.
// Виды типов контента:
// jsonData = "application/json"
// textData = "text/plain"
func (c Client) Get(endpoint string, cType string) (string, error) {
	return c.request(http.MethodGet, endpoint, cType, "")
}
