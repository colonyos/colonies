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
# server on localhost:9000 for pkg/fs. Test packages use per-process databases
# and dynamic ports, so they run in parallel.
test:
	@go test -race ./...

github_test: test

install:
	cp ./bin/colonies /usr/local/bin
	cp ./lib/libcryptolib.so /usr/local/lib
	cp ./lib/libcfslib.so /usr/local/lib

startdb: 
	docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=rFcLGNkgsNtksg6Pgtn9CumL4xXBQ7 --restart unless-stopped timescale/timescaledb:latest-pg16

nukedb:
	@echo "Nuking TimescaleDB containers and volumes..."
	@docker stop $$(docker ps -aq --filter ancestor=timescale/timescaledb:latest-pg16) 2>/dev/null || true
	@docker rm $$(docker ps -aq --filter ancestor=timescale/timescaledb:latest-pg16) 2>/dev/null || true
	@docker volume rm $$(docker volume ls -q --filter dangling=true) 2>/dev/null || true
	@echo "TimescaleDB containers and volumes destroyed"
