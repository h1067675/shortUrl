// Package grpcclient реализует клиента для gRPC сервера
package grpcclient

import (
	// ...
	"context"
	"log"

	pb "github.com/h1067675/shortUrl/internal/grpcserver/proto"
	"github.com/h1067675/shortUrl/internal/logger"
	"go.uber.org/zap"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// GRPCClient описывает структуру настроек сервера
type GRPCClient struct {
	GRPCServerShortener string
	LevelLogging        string
	Client              pb.ShortURLClient
}

// NewGRPCClient создает нового gRPC клиента
func NewGRPCClient() *GRPCClient {
	return &GRPCClient{}
}

// SetNetAddressServer устанавливает адрес сервера для сокращения ссылок
func (g *GRPCClient) SetNetAddressServer(netAddressServer string) *GRPCClient {
	g.GRPCServerShortener = netAddressServer
	return g
}

// StartgRCPClient осуществляет соединение с gRPC сервером
func (g *GRPCClient) StartgRCPClient() {
	// Устанавливаем настройки клиента
	if g.GRPCServerShortener == "" {
		g.GRPCServerShortener = "localhost:8001"
	}

	// устанавливаем соединение с сервером
	conn, err := grpc.NewClient(g.GRPCServerShortener, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Log.Error("Ошибка создания клиента", zap.Error(err))
	}
	// получаем переменную интерфейсного типа UsersClient,
	// через которую будем отправлять сообщения
	g.Client = pb.NewShortURLClient(conn)
}

// Shorten реализует запрос к хэндлеру Shorten
func (g *GRPCClient) Shorten(tokenin string, url string) (tokenout string, body string, statusCode int) {
	link := &pb.Link{Url: url}
	user := &pb.User{Token: tokenin}

	resp, err := g.Client.Shorten(context.Background(), &pb.ShortenRequest{Link: link, User: user})
	if err != nil {
		log.Fatal(err)
	}

	body = resp.Shortlink.Shortlink
	statusCode = int(resp.Status.Status)
	tokenout = resp.User.Token
	return
}

// ShortenJSON реализует запрос к хэндлеру ShortenJSON
func (g *GRPCClient) ShortenJSON(tokenin string, js string) (tokenout string, body string, statusCode int) {
	user := &pb.User{Token: tokenin}

	resp, err := g.Client.ShortenJSON(context.Background(), &pb.ShortenJSONRequest{Json: js, User: user})
	if err != nil {
		log.Fatal(err)
	}

	body = resp.Json
	statusCode = int(resp.Status.Status)
	tokenout = resp.User.Token
	return
}

// ShortenBatchJSON реализует запрос к хэндлеру ShortenBatchJSON
func (g *GRPCClient) ShortenBatchJSON(tokenin string, js string) (tokenout string, body string, statusCode int) {
	user := &pb.User{Token: tokenin}

	resp, err := g.Client.ShortenBatchJSON(context.Background(), &pb.ShortenBatchJSONRequest{Json: js, User: user})
	if err != nil {
		log.Fatal(err)
	}

	body = resp.Json
	statusCode = int(resp.Status.Status)
	tokenout = resp.User.Token
	return
}

// Expand реализует запрос к хэндлеру Expand
func (g *GRPCClient) Expand(url string) (tokenout string, body string, statusCode int) {

	resp, err := g.Client.Expand(context.Background(), &pb.ExpandRequest{Shortlink: &pb.ShortLink{Shortlink: url}})
	if err != nil {
		log.Fatal(err)
	}

	body = resp.Link.Url
	statusCode = int(resp.Status.Status)
	tokenout = resp.User.Token
	return
}

// ExpandUserURLS реализует запрос к хэндлеру ExpandUserURLS
func (g *GRPCClient) ExpandUserURLS(tokenin string, newuser bool) (tokenout string, body string, statusCode int) {
	user := &pb.User{Token: tokenin}

	resp, err := g.Client.ExpandUserURLS(context.Background(), &pb.ExpandUserURLSRequest{User: user})
	if err != nil {
		log.Fatal(err)
	}

	body = resp.Json
	statusCode = int(resp.Status.Status)
	tokenout = resp.User.Token
	return
}

// DeleteUserURLS реализует запрос к хэндлеру DeleteUserURLS
func (g *GRPCClient) DeleteUserURLS(tokenin string, js string) (statusCode int) {
	user := &pb.User{Token: tokenin}

	resp, err := g.Client.DeleteUserURLS(context.Background(), &pb.DeleteUserURLSRequest{User: user, Json: js})
	if err != nil {
		log.Fatal(err)
	}

	statusCode = int(resp.Status.Status)
	return
}

// GetServerStats реализует запрос к хэндлеру GetServerStats
func (g *GRPCClient) GetServerStats() (body string, statusCode int) {
	resp, err := g.Client.GetServerStats(context.Background(), &pb.GetServerStatsRequest{})
	if err != nil {
		log.Fatal(err)
	}

	body = resp.Json
	statusCode = int(resp.Status.Status)
	return
}
