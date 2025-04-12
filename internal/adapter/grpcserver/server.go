package grpcserver

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"time"

	pb "github.com/OksidGen/grpc-file-service/internal/gen"
	"github.com/OksidGen/grpc-file-service/internal/usecase/port"
	"github.com/OksidGen/grpc-file-service/pkg/limiter"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedFileServiceServer
	uc port.FileUseCase
}

var (
	upDownloadSemaphore = limiter.NewLimiter(10)
	listSemaphore       = limiter.NewLimiter(100)
)

func Register(s *grpc.Server, uc port.FileUseCase) {
	pb.RegisterFileServiceServer(s, &server{uc: uc})
}

func (s *server) UploadFile(stream pb.FileService_UploadFileServer) error {
	if err := upDownloadSemaphore.Acquire(stream.Context()); err != nil {
		return status.Error(codes.ResourceExhausted, err.Error())
	}
	defer func() {
		upDownloadSemaphore.Release()
	}()

	time.Sleep(500 * time.Millisecond)

	var data []byte
	var filename string
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		filename = chunk.Filename
		data = append(data, chunk.Data...)
	}
	if err := s.uc.Upload(context.Background(), filename, data); err != nil {
		return err
	}
	return stream.SendAndClose(&pb.UploadResponse{Message: "Загружено"})
}

func (s *server) DownloadFile(req *pb.DownloadRequest, stream pb.FileService_DownloadFileServer) error {
	if err := upDownloadSemaphore.Acquire(stream.Context()); err != nil {
		return status.Error(codes.ResourceExhausted, err.Error())
	}
	defer upDownloadSemaphore.Release()

	time.Sleep(500 * time.Millisecond)

	data, err := s.uc.Download(context.Background(), req.Filename)
	if err != nil {
		return err
	}
	return stream.Send(&pb.DownloadResponse{Data: data})
}

func (s *server) ListFiles(ctx context.Context, _ *pb.ListFilesRequest) (*pb.ListFilesResponse, error) {
	if err := listSemaphore.Acquire(ctx); err != nil {
		return nil, status.Error(codes.ResourceExhausted, err.Error())
	}
	defer listSemaphore.Release()

	time.Sleep(500 * time.Millisecond)

	files, err := s.uc.List(ctx)
	if err != nil {
		return nil, err
	}
	var res []*pb.FileInfo
	for _, f := range files {
		res = append(res, &pb.FileInfo{
			Filename:  f.Filename,
			CreatedAt: f.CreatedAt.Format(time.RFC3339),
			UpdatedAt: f.UpdatedAt.Format(time.RFC3339),
		})
	}
	return &pb.ListFilesResponse{Files: res}, nil
}
