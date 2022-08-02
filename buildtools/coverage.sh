#!/usr/bin/env bash

set -e
echo "" > coverage.txt

go test -race -coverprofile=profile.out -covermode=atomic ./internal/crypto 
if [ -f profile.out ]; then
  cat profile.out >> coverage.txt
  rm profile.out
fi
go test -race -coverprofile=profile.out -covermode=atomic ./pkg/client
if [ -f profile.out ]; then
  cat profile.out >> coverage.txt
  rm profile.out
fi
go test -race -coverprofile=profile.out -covermode=atomic ./pkg/core
if [ -f profile.out ]; then
  cat profile.out >> coverage.txt
  rm profile.out
fi
go test -race -coverprofile=profile.out -covermode=atomic ./pkg/database
if [ -f profile.out ]; then
  cat profile.out >> coverage.txt
  rm profile.out
fi
go test -race -coverprofile=profile.out -covermode=atomic ./pkg/database/postgresql
if [ -f profile.out ]; then
  cat profile.out >> coverage.txt
  rm profile.out
fi
go test -race -coverprofile=profile.out -covermode=atomic ./pkg/rpc
if [ -f profile.out ]; then
  cat profile.out >> coverage.txt
  rm profile.out
fi
go test -race -coverprofile=profile.out -covermode=atomic ./pkg/planner
if [ -f profile.out ]; then
  cat profile.out >> coverage.txt
  rm profile.out
fi
go test -race -coverprofile=profile.out -covermode=atomic ./pkg/planner/basic
if [ -f profile.out ]; then
  cat profile.out >> coverage.txt
  rm profile.out
fi
go test -race -coverprofile=profile.out -covermode=atomic ./pkg/security
if [ -f profile.out ]; then
  cat profile.out >> coverage.txt
  rm profile.out
fi
go test -race -coverprofile=profile.out -covermode=atomic ./pkg/security/crypto
if [ -f profile.out ]; then
  cat profile.out >> coverage.txt
  rm profile.out
fi
go test -race -coverprofile=profile.out -covermode=atomic ./pkg/security/validator
if [ -f profile.out ]; then
  cat profile.out >> coverage.txt
  rm profile.out
fi
go test -race -coverprofile=profile.out -covermode=atomic ./pkg/server
if [ -f profile.out ]; then
  cat profile.out >> coverage.txt
  rm profile.out
fi
go test -race -coverprofile=profile.out -covermode=atomic ./pkg/cluster
if [ -f profile.out ]; then
  cat profile.out >> coverage.txt
  rm profile.out
fi
