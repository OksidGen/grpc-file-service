package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	pb "github.com/OksidGen/grpc-file-service/internal/gen"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const serverAddr = "localhost:50051"

func main() {
	conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(fmt.Sprintf("Не удалось подключиться: %v", err))
	}
	defer func(conn *grpc.ClientConn) {
		_ = conn.Close()
	}(conn)

	client := pb.NewFileServiceClient(conn)

	for {
		fmt.Println("\nВыберите действие:")
		fmt.Println("1. Загрузить файл")
		fmt.Println("2. Скачать файл")
		fmt.Println("3. Посмотреть список файлов")
		fmt.Println("0. Выход")

		var choice int
		fmt.Print("> ")
		_, err := fmt.Scan(&choice)
		if err != nil {
			fmt.Println("Ошибка ввода. Попробуйте снова.")
			continue
		}

		switch choice {
		case 1:
			fmt.Print("Введите путь к файлу для загрузки: ")
			var path string
			if _, err := fmt.Scan(&path); err != nil {
				fmt.Println("Ошибка ввода. Попробуйте снова.")
				continue
			}
			uploadFile(client, path)
		case 2:
			fmt.Print("Введите имя файла на сервере: ")
			var remote string
			if _, err := fmt.Scan(&remote); err != nil {
				fmt.Println("Ошибка ввода. Попробуйте снова.")
				continue
			}
			fmt.Print("Введите имя файла для сохранения: ")
			var local string
			if _, err := fmt.Scan(&local); err != nil {
				fmt.Println("Ошибка ввода. Попробуйте снова.")
				continue
			}
			downloadFile(client, remote, local)
		case 3:
			listFiles(client)
		case 0:
			fmt.Println("Выход.")
			return
		default:
			fmt.Println("Неверный выбор. Попробуйте снова.")
		}
	}
}

func uploadFile(client pb.FileServiceClient, filepath string) {
	file, err := os.Open(filepath)
	if err != nil {
		fmt.Printf("Не удалось открыть файл: %v\n", err)
		return
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	stream, err := client.UploadFile(context.Background())
	if err != nil {
		fmt.Printf("Ошибка при открытии потока загрузки: %v\n", err)
		return
	}

	buf := make([]byte, 1024*64)
	for {
		n, err := file.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Printf("Ошибка при чтении файла: %v\n", err)
			return
		}

		err = stream.Send(&pb.UploadRequest{
			Filename: filepath,
			Data:     buf[:n],
		})
		if err != nil {
			fmt.Printf("Ошибка при отправке чанка: %v\n", err)
			return
		}
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		fmt.Printf("Ошибка при получении ответа: %v\n", err)
		return
	}
	fmt.Printf("Загрузка завершена: %s\n", res.Message)
}

func listFiles(client pb.FileServiceClient) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	resp, err := client.ListFiles(ctx, &pb.ListFilesRequest{})
	if err != nil {
		fmt.Printf("Ошибка при получении списка файлов: %v\n", err)
		return
	}

	fmt.Println("Файлы на сервере:")
	for _, file := range resp.Files {
		fmt.Printf(" - %s | %s | %s\n", file.Filename, file.CreatedAt, file.UpdatedAt)
	}
}

func downloadFile(client pb.FileServiceClient, filename, saveAs string) {
	stream, err := client.DownloadFile(context.Background(), &pb.DownloadRequest{
		Filename: filename,
	})
	if err != nil {
		fmt.Printf("Ошибка при скачивании: %v\n", err)
		return
	}

	outFile, err := os.Create(saveAs)
	if err != nil {
		fmt.Printf("Ошибка при создании файла: %v\n", err)
		return
	}
	defer func(outFile *os.File) {
		_ = outFile.Close()
	}(outFile)

	for {
		resp, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Printf("Ошибка при приёме чанка: %v\n", err)
			return
		}
		_, err = outFile.Write(resp.Data)
		if err != nil {
			fmt.Printf("Ошибка при записи файла: %v\n", err)
			return
		}
	}
	fmt.Printf("Файл сохранён как %s\n", saveAs)
}
