postgres:
	docker run --name simple-bank-db -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -p 5432:5432 -d postgres:17-alpine

createdb:
	docker exec -it simple-bank-db createdb --username=root --owner=root simple_bank

dropdb:
	docker exec -it simple-bank-db dropdb simple_bank

migrateup:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose down

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

server:
	go run cmd/main.go

mock:
	mockgen -package mockdb -destination db/mock/store.go  github.com/hisshihi/simple-bank/db/sqlc Store

.PHONY: createdb dropdb postgres migrateup migratedown sqlc test server mock
