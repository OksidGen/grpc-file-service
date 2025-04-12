package test

import (
	"context"
	"fmt"
	pb "github.com/OksidGen/grpc-file-service/internal/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"os"
)

func setupTestClient() (pb.FileServiceClient, *grpc.ClientConn, func()) {
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Printf("Не удалось подключиться: %v\n", err)
		return nil, nil, func() {}
	}

	client := pb.NewFileServiceClient(conn)

	return client, conn, func() {
		_ = conn.Close()
	}
}

func uploadFile(client pb.FileServiceClient, filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("не удалось открыть файл: %v", err)
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	stream, err := client.UploadFile(context.Background())
	if err != nil {
		return fmt.Errorf("ошибка при открытии потока загрузки: %v", err)
	}

	buf := make([]byte, 1024*64)
	for {
		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			return fmt.Errorf("ошибка при чтении файла: %v", err)
		}
		if err == io.EOF {
			break
		}

		err = stream.Send(&pb.UploadRequest{
			Filename: filepath,
			Data:     buf[:n],
		})
		if err != nil {
			return fmt.Errorf("ошибка при отправке чанка: %v", err)
		}
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		return fmt.Errorf("ошибка при получении ответа: %v", err)
	}

	fmt.Printf("Загрузка завершена: %s\n", res.Message)
	return nil
}

func downloadFile(client pb.FileServiceClient, filename, saveAs string) error {
	stream, err := client.DownloadFile(context.Background(), &pb.DownloadRequest{
		Filename: filename,
	})
	if err != nil {
		return fmt.Errorf("ошибка при скачивании: %v", err)
	}

	outFile, err := os.Create(saveAs)
	if err != nil {
		return fmt.Errorf("ошибка при создании файла: %v", err)
	}
	defer func(outFile *os.File) {
		_ = outFile.Close()
	}(outFile)

	for {
		resp, err := stream.Recv()
		if err != nil && err != io.EOF {
			return fmt.Errorf("ошибка при приёме чанка: %v", err)
		}
		if err == io.EOF {
			break
		}

		_, err = outFile.Write(resp.Data)
		if err != nil {
			return fmt.Errorf("ошибка при записи файла: %v", err)
		}
	}
	fmt.Printf("Файл сохранён как %s\n", saveAs)
	return nil
}
