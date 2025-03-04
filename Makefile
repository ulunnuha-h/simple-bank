postgres:
	docker run --name postgres-db -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=admin123 -d postgres

createdb:
	docker exec -it postgres-db createdb --username=root --owner=root simple_bank

dropdb:
	docker exec -it postgres-db dropdb simple_bank

migrateup:
	migrate -path db/migration/ -database "postgresql://root:admin123@localhost:5432/simple_bank?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration/ -database "postgresql://root:admin123@localhost:5432/simple_bank?sslmode=disable" -verbose down

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

server:
	go run main.go

mock:
	mockgen -package mockdb -destination db/mock/store.go  ./db/sqlc Store

.PHONY: postgres createdb dropdb migrateup migratedown sqlc test mock