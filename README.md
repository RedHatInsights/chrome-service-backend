# chrome-service-backend
Source code repository for chrome backend

# Requirements
Postgre 14
Go 1.18 

## Local Testing

1. There are environment variables that are necessary for the application to start. Please copy the contents within `env.example` and move them over to a new `.env` file at the root of your local app directory. There, set the values of the variables accordingly for your local db configuration. 
2. Run the server by using `go run .` or `go run main.go`
3. To test the service, at the moment, you are able to hit the followind endpoint:
```
GET http://localhost:8000/health
GET http://localhost:8000/api/chrome/v1/hello-world
```

### Headers

To query any endpoint, you will need a `x-rh-identity` header in your request.

You can use this value as an example:

```
eyJpZGVudGl0eSI6eyJ1c2VyIjp7InVzZXJfaWQiOiIxMiJ9fX0=
```