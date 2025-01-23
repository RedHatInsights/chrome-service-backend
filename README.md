# Frontend Operator Configuration Migration

Please follow [the migration docs](./docs/feo-migration-guide.md) to find out more. The complete documentation of the Frontend Operator UI options is in [the starter app docs section](https://github.com/RedHatInsights/frontend-starter-app/blob/master/docs/frontend-operator/index.md).

# Looking for Navigation and All Services Configuration?

> The configuration is handled via Frontend Operator. The static centralized configuration is deprecated.

Follow [this link](./docs/cloud-services-config.md) for docs

- [All services section](./docs/cloud-services-config.md#allservices)

# chrome-service-backend
Source code repository for Chrome backend

# Requirements
Postgres 14
Go 1.21

## Local Testing

1. Run `make env`. There are environment variables that are necessary for the application to start.
   An example of that can be found in `.env.example`, but a basic working env file can be created with `make env`.

2. Run `make infra`. This repo also supports `docker-compose up` for its Postgres server, but `make infra` 
   is recommended to run all needed containers. 

3. Run the server by using `make dev`.
   If you only want to serve the static navigation files, `make dev-static` will serve the needed files from the API.

4. To test the service, at the moment, you are able to hit the following endpoints:

    ```
    GET http://localhost:8000/health
    GET http://localhost:8000/api/chrome-service/v1/hello-world
    ```

### Headers

To query any endpoint, you will need an `x-rh-identity` header in your request.

You can use this value as an example:

```
eyJpZGVudGl0eSI6eyJ1c2VyIjp7InVzZXJfaWQiOiIxMiJ9fX0=
```

### Helpful Make Targets

`make dev` will run the service.

`make infra` will create the DB and Kafka locally.

`make clean` will tear down the database.

`make generate-search-index` will generate the search index file.

`make parse-services` will generate the `services-generated.json` file.
