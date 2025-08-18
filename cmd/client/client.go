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
)

const (
	JSONData = "application/json"
	TextData = "text/plain"
)

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
	logger.Initialize(c.LevelLogging)
	// Устанавливаем настройки приложения по умолчанию
	var conf = configsurl.NewConfig(c.NetAddressServer, c.NetAddressExpand, c.PathStorageFile, c.PathDateBase)
	// Устанавливаем конфигурацию из параметров запуска или из переменных окружения
	conf.Set()
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
func (c *Client) request(method string, endpoint string, cType string, body string) string {
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
		panic(err)
	}
	// в заголовках запроса указываем кодировку
	request.Header.Add("Content-Type", cType)
	// отправляем запрос и получаем ответ
	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	// выводим код ответа
	fmt.Println("Статус-код ", response.Status)
	defer response.Body.Close()
	// читаем поток из тела ответа
	result, err := io.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}
	// и печатаем его
	return string(result)
}

// Post делает POST-запрос к адресной строке endpoint с телом запроса body и типом контента cType.
// Виды типов контента:
// jsonData = "application/json"
// textData = "text/plain"
func (c Client) Post(endpoint string, body string, cType string) string {
	return c.request(http.MethodPost, endpoint, cType, body)
}

// Get делает GET-запрос к адресной строке endpoint и типом контента cType.
// Виды типов контента:
// jsonData = "application/json"
// textData = "text/plain"
func (c Client) Get(endpoint string, cType string) string {
	return c.request(http.MethodGet, endpoint, cType, "")
}
