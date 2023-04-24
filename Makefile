help:
	@echo "Availabe commands:"
	@echo "------------------"
	@echo "migrate		- run database migration"
	@echo "dev 		- run server"
	@echo "test 		- run all tests"
	@echo "database 	- start database with .env vars"
	@echo "clean		- tear down database"
	@echo "dev-static	- serve only the static direcory using simple go server"
	@echo "dev-static-node - serve only the static direcory using simple node server"
	@echo "		  arguments:"
	@echo "			  - port: http server port 'make dev-static-node port=8888'"
	@echo "validate-schema	- validates chrome static JSON schemas"

port?=8000

dev-static:
	go run cmd/static/static.go $(port)


dev-static-node:
	npx http-server . -a :: -p $(port)

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

publish-search-index:
	go run cmd/search/*
