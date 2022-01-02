all: build
.PHONY: all build

IMAGE ?= registry.ice.ri.se/colonies

build:
	@CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o ./bin/colonies ./cmd/main.go
	@go build -buildmode=c-shared -o ./lib/cryptolib.so ./internal/cryptolib/cryptolib.go

docker:
	docker build -t $(IMAGE) .

push-image:
	docker push $(IMAGE)

coverage:
	./buildtools/coverage.sh
	./buildtools/codecov

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
