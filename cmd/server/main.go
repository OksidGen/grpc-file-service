package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/OksidGen/grpc-file-service/internal/adapter/grpcserver"
	"github.com/OksidGen/grpc-file-service/internal/repository"
	"github.com/OksidGen/grpc-file-service/internal/usecase"
	"github.com/OksidGen/grpc-file-service/internal/usecase/port"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	fmt.Println("Запуск сервера...")
	_ = os.MkdirAll("storage", 0755)

	repo := repository.NewLocalFileRepository("storage")
	var fileRepo port.FileRepository = repo
	uc := usecase.NewFileUseCase(fileRepo)
	var fileUseCase port.FileUseCase = uc

	s := grpc.NewServer()
	grpcserver.Register(s, fileUseCase)
	reflection.Register(s)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal(err)
	}
	if err := s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
