all: build
.PHONY: all build

BUILD_IMAGE ?= colonyos/colonies
PUSH_IMAGE ?= colonyos/colonies:v1.5.5

VERSION := $(shell git rev-parse --short HEAD)
BUILDTIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

GOLDFLAGS += -X 'main.BuildVersion=$(VERSION)'
GOLDFLAGS += -X 'main.BuildTime=$(BUILDTIME)'

build:
	@CGO_ENABLED=0 go build -ldflags="-s -w $(GOLDFLAGS)" -o ./bin/colonies ./cmd/main.go
	@go build -buildmode=c-shared -o ./lib/libcryptolib.so ./internal/cryptolib/cryptolib.go
	@GOOS=js GOARCH=wasm go build -o ./lib/libcryptolib.wasm internal/cryptolib.wasm/cryptolib.go

container:
	docker build -t $(BUILD_IMAGE) .

push:
	docker tag $(BUILD_IMAGE) $(PUSH_IMAGE) 
	docker push $(PUSH_IMAGE)

coverage:
	./buildtools/coverage.sh
	./buildtools/codecov

build_cryptolib_ubuntu_2020:
	cd buildtools; ./build_cryptolib_ubuntu.sh 

test:
	@cd tests/reliability; grc go test -v --race
	@cd internal/crypto; grc go test -v --race
	@cd pkg/core; grc go test -v --race
	@cd pkg/database/postgresql; grc go test -v --race
	@cd pkg/rpc; grc go test -v --race
	@cd pkg/security; grc go test -v --race
	@cd pkg/security/crypto; grc go test -v --race
	@cd pkg/security/validator; grc go test -v --race
	@cd pkg/server; grc go test -v --race
	@cd pkg/planner/basic; grc go test -v --race
	@cd pkg/utils; grc go test -v --race
	@cd pkg/cluster; grc go test -v --race
	@cd pkg/cron; grc go test -v --race

github_test: 
	@cd tests/reliability; go test -v --race
	@cd internal/crypto; go test -v --race
	@cd pkg/core; go test -v --race
	@cd pkg/database/postgresql; go test -v --race
	@cd pkg/rpc; go test -v --race
	@cd pkg/security; go test -v --race
	@cd pkg/security/crypto; go test -v --race
	@cd pkg/security/validator; go test -v --race
	@cd pkg/server; go test -v --race
	@cd pkg/planner/basic; go test -v --race
	@cd pkg/utils; go test -v --race
	@cd pkg/cluster; go test -v --race
	@cd pkg/cron; go test -v --race
	@cd pkg/fs; go test -v --race

install:
	cp ./bin/colonies /usr/local/bin
	cp ./lib/libcryptolib.so /usr/local/lib
