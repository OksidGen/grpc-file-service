package repository

import (
	"github.com/OksidGen/grpc-file-service/internal/usecase/port"
	"os"
	//"time"
)

type LocalFileRepository struct {
	dir string
}

func NewLocalFileRepository(dir string) *LocalFileRepository {
	return &LocalFileRepository{dir: dir}
}

func (r *LocalFileRepository) SaveFile(filename string, data []byte) error {
	return os.WriteFile(r.dir+"/"+filename, data, 0644)
}

func (r *LocalFileRepository) ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(r.dir + "/" + filename)
}

func (r *LocalFileRepository) ListFiles() ([]port.FileMeta, error) {
	dirEntries, err := os.ReadDir(r.dir)
	if err != nil {
		return nil, err
	}
	var result []port.FileMeta
	for _, entry := range dirEntries {
		info, _ := entry.Info()
		result = append(result, port.FileMeta{
			Filename:  entry.Name(),
			CreatedAt: info.ModTime(),
			UpdatedAt: info.ModTime(),
		})
	}
	return result, nil
}
