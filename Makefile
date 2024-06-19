help:
	@echo "Availabe commands:"
	@echo "------------------"
	@echo "migrate          	- run database migration"
	@echo "dev              	- run server"
	@echo "test             	- run all tests"
	@echo "database         	- start database with .env vars"
	@echo "kafka            	- start local kafka"
	@echo "unleash          	- start local unleash server"
	@echo "infra            	- start all infrastructure locally (kafka, unleash, and postgres db)"
	@echo "clean            	- tear down database"
	@echo "clean-all        	- tear down all local infrastructure"
	@echo "validate-schema  	- validates chrome static JSON schemas"
	@echo "parse-services 	 	- creates services-generated.json that with filled link refs"
	@echo "dev-static       	- serve only the static direcory using simple go server"
	@echo "dev-static-node  	- serve only the static direcory using simple node server"
	@echo "  arguments:"
	@echo "  - port: http server port 'make dev-static-node port=8888'"
	@echo "audit 		 	- run grype audit on the docker image"
	@echo "generate-search-index 	- generate search index"

port?=8000

dev-static:
	go run cmd/static/static.go $(port)

dev-static-node:
	npx http-server . -a :: -p $(port)

migrate:
	go run cmd/migrate/migrate.go 

database:
	podman-compose up

clean:
	podman-compose down

validate-schema:
	go run cmd/validate/*
	
generate-search-index: export SEARCH_INDEX_WRITE = true

generate-search-index:
	go run cmd/search/*

kafka:
	podman-compose -f local/kafka-compose.yaml up

unleash:
	podman-compose -f local/unleash-compose.yaml up

infra:
	podman-compose -f local/full-stack-compose.yaml up

clean-all:
	podman-compose -f local/full-stack-compose.yaml down

test: seed-unleash
	go test -v  ./... -coverprofile=c.out

coverage:
	go tool cover -html=c.out

seed-unleash:
	go run cmd/unleash/seed.go

dev: seed-unleash
	go run main.go

parse-services:
	go run cmd/services/parseServices.go

audit:
	docker build . -t chrome:audit
	grype chrome:audit --fail-on medium --only-fixed
