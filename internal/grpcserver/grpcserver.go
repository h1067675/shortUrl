package grpcserver

import (
	"context"
	"net"
	"strings"

	"github.com/h1067675/shortUrl/cmd/configsurl"
	pb "github.com/h1067675/shortUrl/internal/grpcserver/proto"
	"github.com/h1067675/shortUrl/internal/logger"
	"go.uber.org/zap"

	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

const (
	keyUserID key = iota
	keyNewUser
	keyUserToken
)

// типы данных
type (
	// Handler связывает с хэндлерами
	Applicator interface {
		Authorization(cookie string, hasToken bool) (token string, userid int, err error)
		CheckDB() (statuscode int)
		Shorten(url string, userid int) (body string, statusCode int)
		ShortenJSON(js []byte, userid int) (body []byte, statusCode int)
		ShortenBatchJSON(js []byte, userid int) (body []byte, statusCode int)
		Expand(shortCode string, userid int) (basedURL string, statusCode int)
		ExpandUserURLS(userid int, newuser bool) (body []byte, statusCode int)
		DeleteUserURLS(js []byte, userid int) (statusCode int)
		GetServerStats(ip net.IP) (body []byte, statusCode int)
	}
	// key необходим для передачи через context
	key int

	// GRPCServer
	GRPCServer struct {
		Server   *grpc.Server
		Listen   net.Listener
		ShortURL ShortURLServer
	}

	// GrcpServer описывает структуру сервера gRCP
	ShortURLServer struct {
		App Applicator
		pb.UnimplementedShortURLServer
	}
)

// Authorization middleware функция осуществляющая авторизацию
func (s *ShortURLServer) Authorization(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	var (
		userid int
	)
	// Проверяем наличие токена
	r, ok := req.(*pb.User)
	if ok {
		token := r.Token
		if token != "" {
			_, userid, err = s.App.Authorization(token, true)
			if err != nil {
				return resp, err
			}
			ctx = context.WithValue(ctx, keyUserID, userid)
			resp, err = handler(ctx, req)
			return resp, err
		}
	}
	token, userid, err := s.App.Authorization("", false)
	if err != nil {
		return resp, err
	}
	ctx = context.WithValue(ctx, keyUserID, userid)
	ctx = context.WithValue(ctx, keyUserToken, token)
	resp, err = handler(ctx, req)
	return resp, err
}

// PingDB фасад к функции PingDB Application
func (s *ShortURLServer) PingDB(ctx context.Context, req *pb.PingDBRequest) (*pb.PingDBResponse, error) {
	response := &pb.PingDBResponse{}

	response.Status = &pb.Status{Status: int32(s.App.CheckDB())}

	return response, nil
}

// Shorten фасад к функции Shorten Application
func (s *ShortURLServer) Shorten(ctx context.Context, req *pb.ShortenRequest) (*pb.ShortenResponse, error) {
	response := &pb.ShortenResponse{}

	body, code := s.App.Shorten(req.Link.Url, ctx.Value(keyUserID).(int))

	response.Status = &pb.Status{Status: int32(code)}
	response.Shortlink = &pb.ShortLink{Shortlink: body}
	response.User = &pb.User{Token: ctx.Value(keyUserToken).(string)}

	return response, nil
}

// ShortenJSON фасад к функции ShortenJSON Application
func (s *ShortURLServer) ShortenJSON(ctx context.Context, req *pb.ShortenJSONRequest) (*pb.ShortenJSONResponse, error) {
	response := &pb.ShortenJSONResponse{}

	body, code := s.App.ShortenJSON([]byte(req.Json), ctx.Value(keyUserID).(int))

	response.Status = &pb.Status{Status: int32(code)}
	response.Json = string(body)
	response.User = &pb.User{Token: ctx.Value(keyUserToken).(string)}

	return response, nil
}

// ShortenBatchJSON фасад к функции ShortenBatchJSON Application
func (s *ShortURLServer) ShortenBatchJSON(ctx context.Context, req *pb.ShortenBatchJSONRequest) (*pb.ShortenBatchJSONResponse, error) {
	response := &pb.ShortenBatchJSONResponse{}

	body, code := s.App.ShortenBatchJSON([]byte(req.Json), ctx.Value(keyUserID).(int))

	response.Status = &pb.Status{Status: int32(code)}
	response.Json = string(body)
	response.User = &pb.User{Token: ctx.Value(keyUserToken).(string)}

	return response, nil
}

// Expand фасад к функции Expand Application
func (s *ShortURLServer) Expand(ctx context.Context, req *pb.ExpandRequest) (*pb.ExpandResponse, error) {
	response := &pb.ExpandResponse{}
	sl := strings.Split(req.Shortlink.Shortlink, "/")
	body, code := s.App.Expand(sl[len(sl)-1], ctx.Value(keyUserID).(int))

	response.Status = &pb.Status{Status: int32(code)}
	response.Link = &pb.Link{Url: string(body)}
	response.User = &pb.User{Token: ctx.Value(keyUserToken).(string)}

	return response, nil
}

// ExpandUserURLS фасад к функции ExpandUserURLS Application
func (s *ShortURLServer) ExpandUserURLS(ctx context.Context, req *pb.ExpandUserURLSRequest) (*pb.ExpandUserURLSResponse, error) {
	response := &pb.ExpandUserURLSResponse{}

	body, code := s.App.ExpandUserURLS(ctx.Value(keyUserID).(int), false)

	response.Status = &pb.Status{Status: int32(code)}
	response.Json = string(body)
	response.User = &pb.User{Token: ctx.Value(keyUserToken).(string)}

	return response, nil
}

// DeleteUserURLS фасад к функции DeleteUserURLS Application
func (s *ShortURLServer) DeleteUserURLS(ctx context.Context, req *pb.DeleteUserURLSRequest) (*pb.DeleteUserURLSResponse, error) {
	response := &pb.DeleteUserURLSResponse{}

	code := s.App.DeleteUserURLS([]byte(req.Json), ctx.Value(keyUserID).(int))

	response.Status = &pb.Status{Status: int32(code)}

	return response, nil
}

// GetServerStats фасад к функции GetServerStats Application
func (s *ShortURLServer) GetServerStats(ctx context.Context, req *pb.GetServerStatsRequest) (*pb.GetServerStatsResponse, error) {
	response := &pb.GetServerStatsResponse{}
	var ip net.IP

	if p, ok := peer.FromContext(ctx); ok {
		ip = net.ParseIP(p.Addr.String())
	}
	body, code := s.App.GetServerStats(ip)

	response.Status = &pb.Status{Status: int32(code)}
	response.Json = string(body)

	return response, nil
}

// New создает gRPC структуру
func New() *GRPCServer {
	return &GRPCServer{}
}

// NewgRPCerver создает gRPC сервер
func (g *GRPCServer) Register(conf configsurl.Config, app Applicator) *GRPCServer {
	var err error
	// определяем порт для сервера
	g.Listen, err = net.Listen("tcp", conf.GRPCServerShortener.String())
	if err != nil {
		logger.Log.Error("Failed to create gRPC server", zap.Error(err))
	}
	// создаём gRPC-сервер без зарегистрированной службы
	s := &ShortURLServer{}
	s.App = app
	g.Server = grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			s.Authorization,
		))
	// регистрируем сервис
	pb.RegisterShortURLServer(g.Server, s)
	return g
}

// StartServergRPC запускает сервер gRPC
func (g *GRPCServer) StartServergRPC(app Applicator) error {
	g.ShortURL.App = app
	go func() {
		if err := g.Server.Serve(g.Listen); err != nil {
			logger.Log.Error("Error to start gRPC server", zap.Error(err))
		}
	}()

	return nil
}
