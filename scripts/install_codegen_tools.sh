#!/usr/bin/env sh

go get -u google.golang.org/protobuf

go install github.com/bufbuild/buf/cmd/buf@v1.5.0
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.1

go install github.com/favadi/protoc-go-inject-tag