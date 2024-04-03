# Include necessary files
include .env

server:
	go run cmd/carpool-btc/main.go

migration:
	migrate create -ext sql -dir db/migration -seq $(NAME)

migrate-up:
ifeq ($(ENVIRONMENT), "PRODUCTION")
	migrate -path db/migration/ -database $(DATABASE_URL) -verbose up
else
	migrate -path db/migration/ -database $(TEST_DATABASE_URL) -verbose up
endif

migrate-down:
ifeq ($(ENVIRONMENT), "PRODUCTION")
	migrate -path db/migration/ -database $(DATABASE_URL) -verbose down
else
	migrate -path db/migration/ -database $(TEST_DATABASE_URL) -verbose down
endif