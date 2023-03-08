help:
	@echo "Availabe commands:"
	@echo "------------------"
	@echo "migrate		- run database migration"
	@echo "dev 				- run server"
	@echo "test 			- run all tests"
	@echo "database 	- start database with .env vars"
	@echo "clean 			- tear down database"

migrate:
	go run cmd/migrate/migrate.go 

dev:
	go run main.go

database:
	podman-compose up

clean:
	podman-compose down
