################################
# STEP 1 build executable binary
################################
FROM registry.access.redhat.com/hi/go:latest-fips-builder AS builder

USER 0

WORKDIR /workspace

# Cache deps before copying source so that we do not need to re-download for every build
COPY go.mod go.sum ./

# Fetch dependencies
RUN go mod download

# Copy source files
COPY api api
COPY cmd cmd
COPY config config
COPY rest rest
COPY static static
COPY main.go main.go
COPY spec spec
COPY Makefile Makefile
COPY widget-dashboard-defaults widget-dashboard-defaults

# Generate static assets
RUN make parse-services
RUN make generate-search-index

# Build all binaries
RUN CGO_ENABLED=1 go build -ldflags "-w -s" -o chrome-service-backend
RUN CGO_ENABLED=1 go build -ldflags "-w -s" -o chrome-migrate cmd/migrate/migrate.go
RUN CGO_ENABLED=1 go build -ldflags "-w -s" -o chrome-search-index cmd/search/publishSearchIndex.go
RUN CGO_ENABLED=1 go build -ldflags "-w -s" -o chrome-fetch-specs cmd/fetchSpecs/fetchSpecs.go

############################
# STEP 2 build a small image
############################
FROM registry.access.redhat.com/hi/go:latest-fips

WORKDIR /

# Setup permissions to allow RDSCA to be written from clowder to container
# https://docs.openshift.com/container-platform/4.11/openshift_images/create-images.html#images-create-guide-openshift_create-images
RUN mkdir -p /app && \
    chgrp -R 0 /app && \
    chmod -R g=u /app
RUN mkdir -p /static && \
    chgrp -R 0 /static && \
    chmod -R g=u /static

COPY --from=builder /workspace/chrome-service-backend /app/chrome-service-backend
COPY --from=builder /workspace/chrome-migrate /usr/bin/
COPY --from=builder /workspace/chrome-search-index /usr/bin/
COPY --from=builder /workspace/chrome-fetch-specs /usr/bin/
# Copy chrome static JSON assets to server binary entry point
COPY --from=builder /workspace/static /static
# Copy widget dashboard defaults to server binary entry point
COPY --from=builder /workspace/widget-dashboard-defaults /widget-dashboard-defaults

USER 1001

EXPOSE 8000
CMD ["/app/chrome-service-backend"]
