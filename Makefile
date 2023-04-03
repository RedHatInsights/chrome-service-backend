help:
	@echo "Availabe commands:"
	@echo "------------------"
	@echo "migrate		- run database migration"
	@echo "dev 		- run server"
	@echo "test 		- run all tests"
	@echo "database 	- start database with .env vars"
	@echo "clean		- tear down database"
	@echo "dev-static	- serve only the static direcory"
	@echo "validate-schema	- validates chrome static JSON schemas"

dev-static:
	go run cmd/static/static.go

migrate:
	go run cmd/migrate/migrate.go 

dev:
	go run main.go

database:
	podman-compose up

clean:
	podman-compose down

validate-schema:
	go run cmd/validate/*
