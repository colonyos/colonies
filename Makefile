all: build
.PHONY: all build

IMAGE ?= colonyos/colonies

build:
	@CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o ./bin/colonies ./cmd/main.go
	@go build -buildmode=c-shared -o ./lib/cryptolib.so ./internal/cryptolib/cryptolib.go
	@GOOS=js GOARCH=wasm go build -o ./lib/cryptolib.wasm internal/cryptolib.wasm/cryptolib.go

docker:
	docker build -t $(IMAGE) .

push-image:
	docker push $(IMAGE)

coverage:
	./buildtools/coverage.sh
	./buildtools/codecov

build_cryptolib_ubuntu_2020:
	cd buildtools; ./build_cryptolib_ubuntu.sh 

test:
	@cd internal/crypto; grc go test -v
	@cd pkg/core; grc go test -v
	@cd pkg/database/postgresql; grc go test -v
	@cd pkg/rpc; grc go test -v
	@cd pkg/security; grc go test -v
	@cd pkg/security/crypto; grc go test -v
	@cd pkg/security/validator; grc go test -v
	@cd pkg/server; grc go test -v
	@cd pkg/scheduler/basic; grc go test -v

github_test:
	@cd internal/crypto; go test -v
	@cd pkg/core; go test -v
	@cd pkg/database/postgresql; go test -v
	@cd pkg/rpc; go test -v
	@cd pkg/security; go test -v
	@cd pkg/security/crypto; go test -v
	@cd pkg/security/validator; go test -v
	@cd pkg/server; go test -v
	@cd pkg/scheduler/basic; go test -v

install:
	cp ./bin/colonies /usr/local/bin
	cp ./lib/cryptolib.so /usr/lib
