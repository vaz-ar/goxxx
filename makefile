SHELL := /bin/bash

VERSION=0.2.0

# -----------------------------------------------------------------------------------------

BUILD_TIME=`date +%FT%T%z`
LDFLAGS=-ldflags "-X main.GlobalVersion=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# -----------------------------------------------------------------------------------------

.DEFAULT_GOAL := build

.PHONY: build install clean test format

build:
	go build $(LDFLAGS) -o build/goxxx goxxx/goxxx.go

install:
	go install $(LDFLAGS) goxxx/goxxx.go

clean:
	if [ -f $(BINARY) ] ; then rm $(BINARY) ; fi
	go clean

test:
	go test -v ./... | sed -e /PASS/s//$$(printf "\033[32mPASS\033[0m")/ -e /FAIL/s//$$(printf "\033[31mFAIL\033[0m")/

format:
	gofmt -s -w ./*.go
	gofmt -s -w ./*/*.go
