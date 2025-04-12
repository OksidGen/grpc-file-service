package port

import (
	"context"
	"time"
)

type FileMeta struct {
	Filename  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type FileUseCase interface {
	Upload(ctx context.Context, filename string, data []byte) error
	Download(ctx context.Context, filename string) ([]byte, error)
	List(ctx context.Context) ([]FileMeta, error)
}

type FileRepository interface {
	SaveFile(filename string, data []byte) error
	ReadFile(filename string) ([]byte, error)
	ListFiles() ([]FileMeta, error)
}
