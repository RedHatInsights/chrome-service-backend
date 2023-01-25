help:
	@echo "Availabe commands:"
	@echo "------------------"
	@echo "migrate		- run database migration"

migrate:
	go run cmd/migrate/migrate.go 
