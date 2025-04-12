# Makefile

# Пути
PROTO_DIR=proto
GEN_DIR=internal/gen
SERVER_DIR=cmd/server
BINARY_NAME=server
OUTPUT_DIR=bin

# Флаги protoc
PROTOC_GEN_GO_ARGS=--go_out=$(GEN_DIR) --go_opt=paths=source_relative
PROTOC_GEN_GRPC_ARGS=--go-grpc_out=$(GEN_DIR) --go-grpc_opt=paths=source_relative

# Задачи
.PHONY: all proto build run clean

all: proto build

proto:
	@echo "Генерация gRPC кода из proto..."
	@mkdir -p $(GEN_DIR)
	protoc -I $(PROTO_DIR) \
		$(PROTOC_GEN_GO_ARGS) \
		$(PROTOC_GEN_GRPC_ARGS) \
		$(PROTO_DIR)/fileservice.proto

build:
	@echo "Сборка бинарника..."
	@mkdir -p $(OUTPUT_DIR)
	go build -o $(OUTPUT_DIR)/$(BINARY_NAME) ./$(SERVER_DIR)

run: build
	@echo "Запуск сервера..."
	./$(OUTPUT_DIR)/$(BINARY_NAME)

clean:
	@echo "Очистка..."
	rm -rf $(OUTPUT_DIR) $(GEN_DIR)