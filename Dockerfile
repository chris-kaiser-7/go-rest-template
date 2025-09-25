# syntax=docker/dockerfile:1

# Build the application from source
FROM golang:1.25 AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

#RUN CGO_ENABLED=0 GOOS=linux go build -o /docker-gs-ping
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags='-s -w' -o=./bin/linux_amd64/api ./cmd/api

# Run the tests in the container
FROM build-stage AS run-test-stage
RUN make test/all

# Deploy the application binary into a lean image
# FROM gcr.io/distroless/base-debian11 AS build-release-stage
FROM debian AS build-release-stage

ENV DB_DSN="postgres://postgres:postgres@localhost:5433/testdb?sslmode=disable"

WORKDIR /

COPY --from=build-stage /app/bin/linux_amd64/api /api

EXPOSE 8080

ENTRYPOINT /api -port=4000 -db-dsn="$DB_DSN" -env=production
