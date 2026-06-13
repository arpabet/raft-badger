.PHONY: all build test proto update

all: build

test:
	go test -race -cover ./...

build: test
	go build -v ./...

# Regenerate protobuf code. Requires protoc and protoc-gen-go on PATH:
#   go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
proto:
	protoc --go_out=. --go_opt=paths=source_relative *.proto

update:
	go get -u ./...
	go mod tidy
