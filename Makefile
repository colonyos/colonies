all: build
.PHONY: all build

BUILD_IMAGE ?= colonyos/colonies
PUSH_IMAGE ?= colonyos/colonies:v1.9.5.beta10

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
	./buildtools/coverage.sh
	./buildtools/codecov

build_cryptolib_ubuntu_2020:
	cd buildtools; ./build_cryptolib_ubuntu.sh 

test:
	@cd tests/reliability; go test -v --race
	@cd internal/crypto; go test -v --race
	@cd pkg/core; go test -v --race
	@cd pkg/database/postgresql; go test -v --race
	@cd pkg/rpc; go test -v --race
	@cd pkg/security; go test -v --race
	@cd pkg/security/crypto; go test -v --race
	@cd pkg/security/validator; go test -v --race
	@cd pkg/backends/gin; go test -v --race
	@cd pkg/backends/grpc; go test -v --race
	@cd pkg/backends/libp2p; go test -v --race
	@cd pkg/client/gin; go test -v --race
	@cd pkg/client/grpc; go test -v --race
	@cd pkg/client/libp2p; go test -v --race
	@cd pkg/server; go test -v --race
	@cd pkg/server/controllers; go test -v --race
	@cd pkg/server/handlers/attribute; go test -v --race
	@cd pkg/server/handlers/colony; go test -v --race
	@cd pkg/server/handlers/cron; go test -v --race
	@cd pkg/server/handlers/executor; go test -v --race
	@cd pkg/server/handlers/file; go test -v --race
	@cd pkg/server/handlers/function; go test -v --race
	@cd pkg/server/handlers/generator; go test -v --race
	@cd pkg/server/handlers/log; go test -v --race
	@cd pkg/server/handlers/process; go test -v --race
	@cd pkg/server/handlers/processgraph; go test -v --race
	@cd pkg/server/handlers/security; go test -v --race
	@cd pkg/server/handlers/server; go test -v --race
	@cd pkg/server/handlers/snapshot; go test -v --race
	@cd pkg/server/handlers/user; go test -v --race
	@cd pkg/server/handlers/realtime; go test -v --race
	@cd pkg/server/utils; go test -v --race
	@cd pkg/scheduler; go test -v --race
	@cd pkg/parsers; go test -v --race
	@cd pkg/utils; go test -v --race
	@cd pkg/cluster; go test -v --race
	@cd pkg/cron; go test -v --race
	@cd pkg/fs; go test -v --race

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
