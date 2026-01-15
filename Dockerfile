FROM registry.access.redhat.com/ubi9/go-toolset:1.25.3-1768393489 AS builder
WORKDIR $GOPATH/src/chrome-service-backend/
# TODO: Use --exclude when stable docker version available
COPY api api
COPY cmd cmd
COPY config config
COPY rest rest
COPY static static
COPY go.mod go.mod
COPY go.sum go.sum
COPY main.go main.go
COPY spec spec
COPY Makefile Makefile
COPY widget-dashboard-defaults widget-dashboard-defaults
ENV GO111MODULE=on
USER root
RUN go get -d -v
RUN make parse-services
RUN make generate-search-index 
RUN CGO_ENABLED=1 go build -o /go/bin/chrome-service-backend
# Build the migration binary.
RUN CGO_ENABLED=1 go build -o /go/bin/chrome-migrate cmd/migrate/migrate.go

# Build the search index binary.
RUN CGO_ENABLED=1 go build -o /go/bin/chrome-search-index cmd/search/publishSearchIndex.go

FROM registry.access.redhat.com/ubi9-minimal:latest

# Setup permissions to allow RDSCA to be written from clowder to container
# https://docs.openshift.com/container-platform/4.11/openshift_images/create-images.html#images-create-guide-openshift_create-images
RUN mkdir -p /app
RUN chgrp -R 0 /app && \
    chmod -R g=u /app
RUN mkdir -p /static
RUN chgrp -R 0 /static && \
    chmod -R g=u /static
COPY --from=builder   /go/bin/chrome-service-backend /app/chrome-service-backend
COPY --from=builder /go/bin/chrome-migrate /usr/bin
COPY --from=builder /go/bin/chrome-search-index /usr/bin
# Copy chrome static JSON assets to server binary entry point
COPY --from=builder $GOPATH/src/chrome-service-backend/static /static
# Copy widget dashboard defaults to server binary entry point
COPY --from=builder $GOPATH/src/chrome-service-backend/widget-dashboard-defaults /widget-dashboard-defaults

ENTRYPOINT ["/app/chrome-service-backend"]
EXPOSE 8000
USER 1001
