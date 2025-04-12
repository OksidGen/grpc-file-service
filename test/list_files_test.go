package test

import (
	"context"
	"testing"

	pb "github.com/OksidGen/grpc-file-service/internal/gen"
	"github.com/stretchr/testify/assert"
)

func TestListFiles(t *testing.T) {
	client, _, cleanup := setupTestClient()
	defer cleanup()

	resp, err := client.ListFiles(context.Background(), &pb.ListFilesRequest{})
	if err != nil {
		t.Fatalf("Ошибка при получении списка файлов: %v", err)
	}

	assert.NotNil(t, resp.Files, "Список файлов пустой")
}
