.PHONY: clean deps test build
.DEFAULT_GOAL := build

DEBUG ?= false

deps:
	go get ./...

update_deps:
	go get -u ./...

test:
	go test

build: deps test _build

_build:
	go build -o sysdeets

clean:
	@rm -f sysdeets
