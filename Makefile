DB_URL=postgresql://postgres:mysecretpassword@localhost:5432/simple_bank?sslmode=disable

postgres:
	docker run --name go-postgres -p 5432:5432 -e POSTGRES_PASSWORD=mysecretpassword -d postgres

createdb:
	docker exec -it go-postgres createdb --username=postgres --owner=postgres simple_bank

dropdb:
	docker exec -it go-postgres dropdb simple_bank

migrateup:
	migrate -path db/migration -database "$(DB_URL)" -verbose up

migratedown:
	migrate -path db/migration -database "$(DB_URL)" -verbose down

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

.PHONY: postgres createdb dropdb test