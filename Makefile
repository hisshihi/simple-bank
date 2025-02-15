postgres:
	docker run --name simplebank -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=simple_bank -p 5432:5432 -d postgres:16

createdb:
	docker exec -it simplebank createdb --username=postgres --owner=postgres simple_bank

dropdb:
	docker exec -it simplebank dropdb simple_bank

migrateup:
	migrate -path db/migration -database "postgresql://postgres:postgres@localhost:5432/simple_bank?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgresql://postgres:postgres@localhost:5432/simple_bank?sslmode=disable" -verbose down

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

server:
	go run main.go

mock:
	mockgen -package mockdb -destination db/mock/store.go github.com/hisshihi/simple-bank-go/db/sqlc Store

.PHONY: postgres createdb dropdb migrateup migratedown sqlc test server mock
