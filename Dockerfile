FROM registry.redhat.io/rhel8/go-toolset:1.18.4-8 AS builder
WORKDIR $GOPATH/src/chrome-service-backend/
COPY . .
ENV GO111MODULE=on
USER root
RUN go get -d -v
RUN make validate-schema
RUN make publish-search-index-dry-run
RUN CGO_ENABLED=1 go build -o /go/bin/chrome-service-backend

# Build the migration binary.
RUN CGO_ENABLED=1 go build -o /go/bin/chrome-migrate cmd/migrate/migrate.go

# Build the search index binary.
RUN CGO_ENABLED=1 go build -o /go/bin/chrome-search-index cmd/search/publishSearchIndex.go

FROM registry.redhat.io/ubi8-minimal:latest

# Setup permissions to allow RDSCA to be written from clowder to container
# https://docs.openshift.com/container-platform/4.11/openshift_images/create-images.html#images-create-guide-openshift_create-images
RUN mkdir -p /app
RUN chgrp -R 0 /app && \
    chmod -R g=u /app
COPY --from=builder   /go/bin/chrome-service-backend /app/chrome-service-backend
COPY --from=builder /go/bin/chrome-migrate /usr/bin
COPY --from=builder /go/bin/chrome-search-index /usr/bin
# Copy chrome static JSON assets to server binary entry point
COPY --from=builder $GOPATH/src/chrome-service-backend/static /static

ENTRYPOINT ["/app/chrome-service-backend"]
EXPOSE 8000
USER 1001
