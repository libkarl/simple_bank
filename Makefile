postgres: 
	docker run -p 5433:5432 --name postgres14New --network bank-network -e POSTGRES_USER=root  -e POSTGRES_PASSWORD=secret -d postgres
createdb: 
	docker exec -it postgres14New createdb --username=root --owner=root simple_bank

dropdb:
	docker exec -it postgres14New dropdb --username=root --owner=root simple_bank 

migrateup:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5433/simple_bank?sslmode=disable" -verbose up

migrateup1:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5433/simple_bank?sslmode=disable" -verbose up 1

migratedown:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5433/simple_bank?sslmode=disable" -verbose down
migratedown1:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5433/simple_bank?sslmode=disable" -verbose down 1

sqlc: 
	./sqlc generate

test: 
	go test -v -cover ./...

server:
	go run main.go

mock: 
	mockgen -package mockdb -destination db/mock/store.go github.com/karlib/simple_bank/db/sqlc Store

.PHONY: postgres createdb dropdb migrateup migratedown, sqlc, test, server, mock, migrateup1, migratedown1