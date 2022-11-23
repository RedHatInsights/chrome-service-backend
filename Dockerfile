FROM registry.redhat.io/rhel8/go-toolset:1.18.4-8 AS builder
WORKDIR $GOPATH/src/chrome-service-backend/
COPY . .
ENV GO111MODULE=on
USER root
RUN go get -d -v
RUN CGO_ENABLED=0 go build -o /go/bin/chrome-service-backend

FROM registry.redhat.io/ubi8-minimal:latest

COPY --from=builder /go/bin/chrome-service-backend /usr/bin

USER 1001


CMD ["chrome-service-backend"]
EXPOSE 8000
