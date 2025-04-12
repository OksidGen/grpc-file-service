package test

import (
	"context"
	"fmt"
	pb "github.com/OksidGen/grpc-file-service/internal/gen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
	"google.golang.org/grpc"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

const (
	upDownloadLimit = 10
	listLimit       = 100
	extra           = 5
)

func TestConcurrentUploadConnectionsLimit(t *testing.T) {
	testFiles := make([]string, upDownloadLimit+extra)
	for i := 0; i < upDownloadLimit+extra; i++ {
		tempFile := fmt.Sprintf("./test_file_%d.jpg", i)
		err := os.WriteFile(tempFile, []byte("test image content"), 0644)
		require.NoError(t, err)
		testFiles[i] = tempFile
		defer func(name string) {
			_ = os.Remove(name)
		}(tempFile)
	}

	startCh := make(chan struct{}) // Для одновременного старта
	blockCh := make(chan struct{}) // Для блокировки успешных загрузок

	var wg sync.WaitGroup
	successCount := atomic.NewInt32(0)
	failedCount := atomic.NewInt32(0)

	for i := 0; i < upDownloadLimit+extra; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			client, _, cleanup := setupTestClient()
			defer cleanup()

			<-startCh

			err := uploadFile(client, testFiles[id])
			if err != nil {
				if isLimitExceededError(err) {
					failedCount.Inc()
					t.Logf("Запрос %d: лимит превышен (ожидаемо)", id)
				} else {
					t.Errorf("Запрос %d: неожиданная ошибка: %v", id, err)
				}
				return
			}

			successCount.Inc()
			t.Logf("Запрос %d: успешно", id)

			<-blockCh
		}(i)
	}

	close(startCh)

	time.Sleep(3 * time.Second)

	assert.Equal(t, int32(upDownloadLimit), successCount.Load(),
		"Количество успешных загрузок должно соответствовать лимиту")

	assert.Equal(t, int32(extra), failedCount.Load(),
		"Количество отклоненных загрузок должно соответствовать превышению лимита")

	close(blockCh)
	wg.Wait()

	t.Logf("Успешных загрузок: %d, отклоненных: %d",
		successCount.Load(), failedCount.Load())
}

func TestConcurrentListFilesConnectionsLimit(t *testing.T) {
	total := listLimit + extra*4

	clients := make([]pb.FileServiceClient, total)
	cleanups := make([]func(), total)

	for i := 0; i < total; i++ {
		client, conn, cleanup := setupTestClient()
		clients[i] = client
		cleanups[i] = cleanup
		defer cleanup()
		defer func(conn *grpc.ClientConn) {
			_ = conn.Close()
		}(conn)
	}

	startCh := make(chan struct{}) // для одновременного старта
	blockCh := make(chan struct{}) // для блокировки успешных запросов

	var wg sync.WaitGroup
	successCount := atomic.NewInt32(0)
	failedCount := atomic.NewInt32(0)

	for i := 0; i < total; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			<-startCh

			_, err := clients[id].ListFiles(context.Background(), &pb.ListFilesRequest{})

			if err != nil {
				if isLimitExceededError(err) {
					failedCount.Inc()
					t.Logf("Запрос %d: лимит превышен (ожидаемо)", id)
				} else {
					t.Errorf("Запрос %d: неожиданная ошибка: %v", id, err)
				}
				return
			}

			successCount.Inc()
			t.Logf("Запрос %d: успешно", id)

			<-blockCh
		}(i)
	}

	close(startCh)

	time.Sleep(3 * time.Second)

	assert.Equal(t, int32(listLimit), successCount.Load(),
		"Количество успешных запросов должно соответствовать лимиту")

	assert.Equal(t, int32(extra*4), failedCount.Load(),
		"Количество отклоненных запросов должно соответствовать превышению лимита")

	close(blockCh)
	wg.Wait()

	t.Logf("Успешных запросов: %d, отклоненных: %d",
		successCount.Load(), failedCount.Load())
}

func isLimitExceededError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "too many requests")
}
