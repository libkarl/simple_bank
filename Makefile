postgres: 
	docker run -p 5433:5432 --name postgres14New  -e POSTGRES_USER=root  -e POSTGRES_PASSWORD=secret -d postgres:14-alpine
createdb: 
	docker exec -it postgres14New createdb --username=root --owner=root simple_bank

dropdb:
	docker exec -it postgres14New dropdb --username=root --owner=root simple_bank 

migrateup:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5433/simple_bank?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5433/simple_bank?sslmode=disable" -verbose down

sqlc: 
	sqlc generate

.PHONY: postgres createdb dropdb migrateup migratedown, sqlc