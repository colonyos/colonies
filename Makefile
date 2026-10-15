all: build
.PHONY: all build container container-multiplatform container-multiplatform-push push coverage test test-all test-compat github_test install startdb nukedb

BUILD_IMAGE ?= colonyos/colonies
PUSH_IMAGE ?= colonyos/colonies:v1.9.10

VERSION := $(shell git rev-parse --short HEAD)
BUILDTIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

GOLDFLAGS += -X 'main.BuildVersion=$(VERSION)'
GOLDFLAGS += -X 'main.BuildTime=$(BUILDTIME)'

build:
	@CGO_ENABLED=0 go build -ldflags="-s -w $(GOLDFLAGS)" -o ./bin/colonies ./cmd/main.go
	@go build -buildmode=c-shared -o ./lib/libcryptolib.so ./internal/cryptolib/cryptolib.go
	@go build -buildmode=c-shared -o ./lib/libcfslib.so ./internal/cfslib/cfslib.go
	@GOOS=js GOARCH=wasm go build -o ./lib/libcryptolib.wasm internal/cryptolib.wasm/cryptolib.go

container:
	@echo "Building container for local architecture..."
	docker build --build-arg VERSION=$(VERSION) --build-arg BUILDTIME=$(BUILDTIME) -t $(BUILD_IMAGE) .

container-multiplatform:
	@echo "Building multiplatform container (amd64, arm64)..."
	docker buildx build --platform linux/amd64,linux/arm64 --build-arg VERSION=$(VERSION) --build-arg BUILDTIME=$(BUILDTIME) -t $(BUILD_IMAGE) .

container-multiplatform-push:
	@echo "Building and pushing multiplatform container (amd64, arm64)..."
	docker buildx build --platform linux/amd64,linux/arm64 --build-arg VERSION=$(VERSION) --build-arg BUILDTIME=$(BUILDTIME) -t $(BUILD_IMAGE) -t $(PUSH_IMAGE) --push .

push:
	docker tag $(BUILD_IMAGE) $(PUSH_IMAGE)
	docker push $(BUILD_IMAGE)
	docker push $(PUSH_IMAGE)

coverage:
	@go test -coverprofile=coverage.txt -covermode=atomic ./...

build_cryptolib_ubuntu_2020:
	cd buildtools; ./build_cryptolib_ubuntu.sh 

# Runs all tests: needs Postgres on localhost:5432 (make startdb) and an S3
# server on localhost:9000 for pkg/fs (make starts3, then export the AWS_S3_*
# variables printed by that target). Test packages use per-process databases
# and dynamic ports, so they run in parallel.
test:
	@go test -race ./...

github_test: test

install:
	cp ./bin/colonies /usr/local/bin
	cp ./lib/libcryptolib.so /usr/local/lib
	cp ./lib/libcfslib.so /usr/local/lib

# max_connections is raised from the default 100 because go test runs many
# packages in parallel, each with its own database, connection pool and
# TimescaleDB background worker
startdb:
	docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=rFcLGNkgsNtksg6Pgtn9CumL4xXBQ7 --restart unless-stopped timescale/timescaledb:latest-pg16 -c max_connections=300

nukedb:
	@echo "Nuking TimescaleDB containers and volumes..."
	@docker stop $$(docker ps -aq --filter ancestor=timescale/timescaledb:latest-pg16) 2>/dev/null || true
	@docker rm $$(docker ps -aq --filter ancestor=timescale/timescaledb:latest-pg16) 2>/dev/null || true
	@docker volume rm $$(docker volume ls -q --filter dangling=true) 2>/dev/null || true
	@echo "TimescaleDB containers and volumes destroyed"

# S3 server for pkg/fs tests. SeaweedFS mini mode seeds the credentials and
# bucket from the environment and serves S3 on 8333, mapped to 9000 here.
starts3:
	docker run -d -p 9000:8333 -e AWS_ACCESS_KEY_ID=testaccesskey -e AWS_SECRET_ACCESS_KEY=testsecretkey -e S3_BUCKET=test --restart unless-stopped chrislusf/seaweedfs:4.47
	@echo "Export before running tests:"
	@echo "  export AWS_S3_ENDPOINT=localhost:9000 AWS_S3_ACCESSKEY=testaccesskey AWS_S3_SECRETKEY=testsecretkey AWS_S3_TLS=false AWS_S3_SKIPVERIFY=false AWS_S3_BUCKET=test"

nukes3:
	@echo "Nuking SeaweedFS containers..."
	@docker stop $$(docker ps -aq --filter ancestor=chrislusf/seaweedfs:4.47) 2>/dev/null || true
	@docker rm $$(docker ps -aq --filter ancestor=chrislusf/seaweedfs:4.47) 2>/dev/null || true
	@echo "SeaweedFS containers destroyed"
