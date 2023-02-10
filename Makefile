postgres:
	docker run -d -p 5432:5432 --name postgres -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret postgres:14.2-alpine

dbshell:
	docker exec -it postgres psql -U root -d nextcrm

createdb:
	docker exec -it postgres createdb --username=root --owner=root nextcrm

dropdb:
	docker exec -it postgres dropdb --username=root nextcrm

dbmigrateup:
	go run migrateup/main.go

dbmigratedown:
	go run migratedown/main.go

run:
	go run main.go

build:
	go build -o main main.go

.PHONY:
	postgres
	dbshell
	createdb
	dropdb
	migrateup
	migratedown
	run
	build