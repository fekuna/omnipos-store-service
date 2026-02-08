PROJECT_NAME := omnipos-store-service
PROTO_PATH := ../omnipos-proto
DB_URL := postgres://postgres:postgres@localhost:5432/omnipos_store_db?sslmode=disable
MIGRATE_PATH := migrations

.PHONY: build run migrate_up migrate_down create_migration

build:
	go build -v -o bin/server ./cmd/grpc

run: build
	./bin/server

migrate_up:
	migrate -path $(MIGRATE_PATH) -database "$(DB_URL)" up

migrate_down:
	migrate -path $(MIGRATE_PATH) -database "$(DB_URL)" down

create_migration:
	migrate create -ext sql -dir $(MIGRATE_PATH) -seq $(name)

proto:
	cd $(PROTO_PATH) && buf generate
