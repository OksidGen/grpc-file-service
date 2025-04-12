package limiter

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sync"

	"google.golang.org/grpc"
)

var (
	uploadDownloadSem = NewSemaphore(10)
	listSem           = NewSemaphore(100)
)

type Semaphore struct {
	ch chan struct{}
	mu sync.Mutex
}

func NewSemaphore(limit int) *Semaphore {
	return &Semaphore{ch: make(chan struct{}, limit)}
}

func (s *Semaphore) Acquire() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.ch) >= cap(s.ch) {
		return false
	}

	s.ch <- struct{}{}
	return true
}

func (s *Semaphore) Release() {
	s.mu.Lock()
	defer s.mu.Unlock()

	select {
	case <-s.ch:
	default:
	}
}

func NewUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		sem := selectSemaphore(info.FullMethod)
		if !sem.Acquire() {
			return nil, status.Error(codes.Unavailable, "too many concurrent requests")
		}
		defer sem.Release()
		return handler(ctx, req)
	}
}

func NewStreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		sem := selectSemaphore(info.FullMethod)
		if !sem.Acquire() {
			return status.Error(codes.Unavailable, "too many concurrent streams")
		}
		defer sem.Release()
		return handler(srv, ss)
	}
}

func selectSemaphore(method string) *Semaphore {
	switch method {
	case "/fileservice.FileService/UploadFile":
		return uploadDownloadSem
	case "/fileservice.FileService/DownloadFile":
		return uploadDownloadSem
	case "/fileservice.FileService/ListFiles":
		return listSem
	default:
		return listSem
	}
}
