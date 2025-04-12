package test

import (
	"os"
	"testing"
)

func TestDownloadFile(t *testing.T) {
	client, _, cleanup := setupTestClient()
	defer cleanup()

	remoteFilename := "test_file.jpg"
	localFilename := "downloaded_test_file.jpg"

	err := downloadFile(client, remoteFilename, localFilename)
	if err != nil {
		t.Fatalf("Ошибка при скачивании файла: %v", err)
	}

	_, err = os.Stat(localFilename)
	if err != nil {
		t.Fatalf("Не удалось найти скачанный файл: %v", err)
	}

	defer func(name string) {
		_ = os.Remove(name)
	}(localFilename)
}
