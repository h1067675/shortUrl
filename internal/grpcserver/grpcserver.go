package grpcserver

import (
	"context"
	"fmt"
	pb "grpcserver/proto"
	"log"
	"net"

	"google.golang.org/grpc"
)

// GrcpServer описывает структуру сервера gRCP
type GrpcServer struct {
	pb.UnimplementedShortUrlServer
}

// PingDB проверяет наличие базы данных.
func (s *GrpcServer) PingDB(ctx context.Context, req *pb.PingDBRequest) (*pb.PingDBResponse, error) {
	var response pb.PingDBResponse
	response.Status.Status = int32(200)

	return &response, nil
}

func StartgRCPServer() {
	// определяем порт для сервера
	listen, err := net.Listen("tcp", ":3200")
	if err != nil {
		log.Fatal(err)
	}
	// создаём gRPC-сервер без зарегистрированной службы
	s := grpc.NewServer()
	// регистрируем сервис
	// pb.RegisterUsersServer(s, &UsersServer{})

	fmt.Println("Сервер gRPC начал работу")
	// получаем запрос gRPC
	if err := s.Serve(listen); err != nil {
		log.Fatal(err)
	}
}
