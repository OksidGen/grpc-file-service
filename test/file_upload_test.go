package test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileUpload(t *testing.T) {
	client, _, cleanup := setupTestClient()
	defer cleanup()

	tempFile := "./test_file.jpg"
	err := os.WriteFile(tempFile, []byte("test image content"), 0644)
	if err != nil {
		t.Fatalf("Не удалось создать файл для теста: %v", err)
	}
	defer func(name string) {
		_ = os.Remove(name)
	}(tempFile)

	err = uploadFile(client, tempFile)
	if err != nil {
		t.Fatalf("Ошибка при загрузке файла: %v", err)
	}

	fileInfo, err := os.Stat(tempFile)
	if err != nil {
		t.Fatalf("Не удалось получить информацию о файле: %v", err)
	}

	assert.True(t, fileInfo.Size() > 0, "Файл не был загружен")
}
