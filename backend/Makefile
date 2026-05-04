.PHONY: help run build test migrate-up migrate-down clean

help:
	@echo "FinVue Backend Makefile"
	@echo "======================="
	@echo "make run          - Запуск сервера"
	@echo "make build        - Сборка бинарного файла"
	@echo "make test         - Запуск тестов"
	@echo "make migrate-up   - Применение миграций"
	@echo "make migrate-down - Откат миграций"
	@echo "make clean        - Очистка бинарных файлов"

run:
	go run cmd/server/main.go

build:
	go build -o bin/server cmd/server/main.go

test:
	go test -v ./...

migrate-up:
	migrate -path migrations -database "postgresql://finvue:finvue_secret@localhost:5432/finvue?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgresql://finvue:finvue_secret@localhost:5432/finvue?sslmode=disable" down

clean:
	rm -rf bin/