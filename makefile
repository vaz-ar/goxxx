VERSION=0.0.3
BINARY=goxxx

# -----------------------------------------------------------------------------------------

# This will only work while go version is < 2
GO_GTE_15=$(shell [ $$(go version | sed -r 's/^.*[0-9]\.([0-9])\.[0-9].*$$/\1/') -ge 5 ] && echo true || echo false)

ifeq ($(GO_GTE_15), true)
	LDFLAGS=-ldflags "-X main.GlobalVersion=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"
else
	LDFLAGS=-ldflags "-X main.GlobalVersion $(VERSION) -X main.BuildTime $(BUILD_TIME)"
endif

BUILD_TIME=`date +%FT%T%z`
SOURCEDIR=.
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')

# -----------------------------------------------------------------------------------------

.DEFAULT_GOAL := $(BINARY)

$(BINARY): $(SOURCES)
	go build $(LDFLAGS) -o $(BINARY) goxxx.go

.PHONY: install
install: $(SOURCES)
	go install $(LDFLAGS) goxxx.go

.PHONY: clean
clean:
	if [ -f $(BINARY) ] ; then rm $(BINARY) ; fi
	go clean

.PHONY: test
test:
	go test -v ./... | sed -e /PASS/s//$$(printf "\033[32mPASS\033[0m")/ -e /FAIL/s//$$(printf "\033[31mFAIL\033[0m")/


.PHONY: format
format:
	gofmt -s -w ./*.go
	gofmt -s -w ./*/*.go
