// Package gprcclient test
package grpcclient

import (
	"fmt"
	"testing"
	"time"

	"github.com/h1067675/shortUrl/cmd/configsurl"
	"github.com/h1067675/shortUrl/cmd/storage"
	"github.com/h1067675/shortUrl/internal/app"
	"github.com/h1067675/shortUrl/internal/grpcserver"
	"github.com/h1067675/shortUrl/internal/logger"
	"github.com/h1067675/shortUrl/internal/router"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// Client описывает структуру настроек сервера
type (
	test struct {
		name  string
		body  string
		token string
		want  want
	}
	want struct {
		body  string
		token string
		code  int
	}
	Server struct {
		NetAddressServer    string
		NetAddressExpand    string
		GRPCServerShortener string
		GRPCServerExpand    string
		PathStorageFile     string
		PathDateBase        string
		LevelLogging        string
	}
)

// NewClient создает нового клиента
func NewServer() *Server {
	return &Server{}
}

// SetNetAddressServer устанавливает адрес сервера для сокращения ссылок
func (c *Server) SetNetAddressServer(netAddressServer string) *Server {
	c.NetAddressServer = netAddressServer
	return c
}

// SetNetAddressExpans устанавливает адрес сервера извлечения исходного адреса из короткой ссылки
func (c *Server) SetNetAddressExpans(netAddressExpand string) *Server {
	c.NetAddressExpand = netAddressExpand
	return c
}

// SetPathStorageFile устанавливает путь для хранения резервного файла данных
func (c *Server) SetPathStorageFile(pathStorageFile string) *Server {
	c.PathStorageFile = pathStorageFile
	return c
}

// SetPathDateBase устанавливает путь доступа к базе данных
func (c *Server) SetPathDateBase(pathDateBase string) *Server {
	c.PathDateBase = pathDateBase
	return c
}

// SetLevelLogging устанавливает уровень сбора логов
func (c *Server) SetLevelLogging(levelLogging string) *Server {
	c.LevelLogging = levelLogging
	return c
}

// StartServer запускает сервер сокращения ссылок
func (c *Server) StartServer() {
	if c.NetAddressServer == "" {
		c.NetAddressServer = "localhost:8080"
	}
	if c.NetAddressExpand == "" {
		c.NetAddressExpand = "localhost:8080"
	}
	if c.GRPCServerShortener == "" {
		c.GRPCServerShortener = "localhost:8001"
	}
	if c.GRPCServerExpand == "" {
		c.GRPCServerExpand = "localhost:8001"
	}
	if c.PathStorageFile == "" {
		c.PathStorageFile = "./storage.json"
	}
	if c.PathDateBase == "" {
		c.PathDateBase = "host=127.0.0.1 port=5432 dbname=postgres user=postgres password=12345678 connect_timeout=10 sslmode=prefer"
	}
	if c.LevelLogging == "" {
		c.LevelLogging = "debug"
	}
	// Инициализируем логгер
	err := logger.Initialize(c.LevelLogging)
	if err != nil {
		fmt.Print(err)
	}
	// Устанавливаем настройки приложения по умолчанию
	conf, err := configsurl.NewConfig(c.NetAddressServer, c.NetAddressExpand, c.GRPCServerShortener, c.GRPCServerExpand, c.PathStorageFile, c.PathDateBase)
	if err != nil {
		logger.Log.Debug("Уrrors when configuring the server", zap.String("Error", err.Error()))
	}

	// Создаем хранилище данных
	var st = storage.NewStorage(conf.DatabaseDSN.String())
	// Если соединение с базой данных не установлено или не получилось создать таблицу, то загружаем ссылки из файла
	if !st.DB.Connected && conf.FileStoragePath.Path != "" {
		st.RestoreFromfile(conf.FileStoragePath.Path)
	}

	// Создаем соединенbz и маршрутизацию
	serverHTTP := router.New()
	servergRPC := grpcserver.New()

	// запускаем бизнес-логику и помещвем в нее переменные хранения, конфигурации и маршрутизации
	var application app.Application
	application.New(st, conf, serverHTTP, servergRPC)

	// Запускаем сервер
	go application.StartServers()
}

func Test_shortenHandler(t *testing.T) {
	tests := []test{
		{
			name:  "test shorten #1",
			body:  "http://ys.ru",
			token: "",
			want: want{
				code:  201,
				body:  "",
				token: "",
			},
		},
		{
			name:  "test shorten #2",
			body:  "http://ys.ru",
			token: "",
			want: want{
				code:  409,
				body:  "+",
				token: "+",
			},
		},
	}

	c := NewGRPCClient()
	c.StartgRCPClient()
	s := NewServer()
	s.StartServer()

	time.Sleep(time.Second * 2)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// resp, _ := c.Client.Shorten(context.Background(), &pb.ShortenRequest{Link: &pb.Link{Url: test.body}})
			// fmt.Print(resp)
			// assert.Equal(t, resp.Status, test.want.code)
			token, _, code := c.Shorten(test.token, test.body)
			assert.Equal(t, code, test.want.code)
			assert.NotEmpty(t, token)
			assert.NotEmpty(t, test.body)
		})
	}
}
