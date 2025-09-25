# Include env variables

include .envrc
POSTGRES_TEST_DSN ?= 'postgres://postgres:postgres@localhost:5433/testdb?sslmode=disable'


# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^//'

.PHONY: confirm
confirm:
	@echo 'Are you sure? [y/N]' && read ans && [ $${ans:-N} = y ]

# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## run/api: run the cmd/api application
.PHONY: run/api
run/api:
	./hack/test/docker-postgres.sh

	@go run ./cmd/api -db-dsn=${POSTGRES_TEST_DSN}

## db/psql: connect to the database using psql
.PHONY: db/psql
db/psql:
	psql ${DB_DSN}

## db/migrations/new name=$1: create a new database migration
.PHONY: db/migrations/new
db/migrations/new:
	@echo 'Creating migration files for ${name}'
	migrate create -seq -ext=.sql -dir=./migrations ${name}

## db/migrations/up: apply all up database migrations
.PHONY: db/migrations/up
db/migrations/up: confirm
	@echo 'Running up migrations...'
	migrate -path="./migrations" -database ${DB_DSN} up

## db/migrations/down: apply all down database migrations
.PHONY: db/migrations/down
db/migrations/down:
	@echo 'Running down migrations...'
	migrate -path="./migrations" -database ${DB_DSN} down

# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

## test/data: test internal data
.PHONY: test/data
test/data:
	./hack/test/docker-postgres.sh

	@go test ./internal/data -v -dsn=${POSTGRES_TEST_DSN}

## test/api: test api
.PHONY: test/api
test/api:
	./hack/test/docker-postgres.sh

	@go test ./cmd/api -v -dsn=${POSTGRES_TEST_DSN}

## test/all: test all
.PHONY: test/all
test/all:
	./hack/test/docker-postgres.sh

	@go test ./internal/data -v -dsn=${POSTGRES_TEST_DSN}
	@go test ./cmd/api -v -dsn=${POSTGRES_TEST_DSN}



## audit: tidy dependencies and format, vet, and test all code
.PHONY: audit
audit: vendor
	@echo 'Formatting code...'
	go fmt ./...
	@echo 'Vetting code...'
	go vet ./...
	staticcheck ./...
	@echo 'Running tests...'
	go test -race -vet=off ./...

## vendor: tidy and vendor dependencies
.PHONY: vendor
vendor:
	@echo 'Tidying and verifying module dependencies'
	go mod tidy
	go mod verify
	@echo 'Vendoring dependencies'
	go mod vendor

# ==================================================================================== #
# BUILD
# ==================================================================================== #

## build/api: build the cmd/api application
.PHONY: build/api
build/api:
	@echo 'Building cmd/api...'
	go build -ldflags '-s -w' -o ./bin/api ./cmd/api
	GOOS=linux GOARCH=amd64 go build -ldflags='-s -w' -o=./bin/linux_amd64/api ./cmd/api

# ==================================================================================== #
# PRODUCTION
# ==================================================================================== #

PRODUCTION_HOST_IP = '144.126.210.226'

## production/connect: connect to the production server
.PHONY: production/connect
production/connect:
	ssh greenlight@${PRODUCTION_HOST_IP}

## production/deploy/api: deploy the api to production
.PHONY: production/deploy/api
production/deploy/api:
	rsync -P ./bin/linux_amd64/api greenlight@${PRODUCTION_HOST_IP}:~
	rsync -rP --delete ./migrations greenlight@${PRODUCTION_HOST_IP}:~
	rsync -P ./remote/production/api.service greenlight@${PRODUCTION_HOST_IP}:~
	rsync -P ./remote/production/Caddyfile greenlight@${PRODUCTION_HOST_IP}:~
	ssh -t greenlight@${PRODUCTION_HOST_IP} '\
		migrate -path ~/migrations -database $$DB_DSN up \
        && sudo mv ~/api.service /etc/systemd/system/ \
        && sudo systemctl enable api \
        && sudo systemctl restart api \
        && sudo mv ~/Caddyfile /etc/caddy/ \
        && sudo systemctl reload caddy \
      '
