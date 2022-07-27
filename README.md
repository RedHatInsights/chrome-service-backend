# chrome-service-backend
Source code repository for chrome backend

## Local Testing

1. There are environment variables that are necessary for the application to start. Please copy the contents within `env.example` and move them over to a new `.env` file at the root of your local app directory. There, set the values of the variables accordingly for your local db configuration. 
2. Run the server by using `go run .` or `go run main.go`
3. To test the service, at the moment, you are able to hit the followind endpoint:
```
GET http://localhost:8000/health
GET http://localhost:8000/api/chrome/v1/hello-world
```
