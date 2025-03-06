postgres:
	docker run --name simplebank-postgres -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -e POSTGRES_DB=simple_bank -p 5432:5432 -d postgres:latest

createdb:
	docker exec -it simplebank-postgres createdb --username=root --owner=root simple_bank

dropdb:
	docker exec -it simplebank-postgres dropdb --username=root simple_bank

migrateup:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose down

sqlc:
	sqlc generate

.PHONY: postgres createdb dropdb migrateup migratedown
