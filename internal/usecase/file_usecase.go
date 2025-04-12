package usecase

import (
	"context"
	"github.com/OksidGen/grpc-file-service/internal/usecase/port"
)

type fileUseCase struct {
	repo port.FileRepository
}

func NewFileUseCase(repo port.FileRepository) port.FileUseCase {
	return &fileUseCase{repo: repo}
}

func (f *fileUseCase) Upload(ctx context.Context, filename string, data []byte) error {
	return f.repo.SaveFile(filename, data)
}

func (f *fileUseCase) Download(ctx context.Context, filename string) ([]byte, error) {
	return f.repo.ReadFile(filename)
}

func (f *fileUseCase) List(ctx context.Context) ([]port.FileMeta, error) {
	return f.repo.ListFiles()
}
