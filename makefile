VERSION=0.0.3
BINARY=goxxx

BUILD_TIME=`date +%FT%T%z`
SOURCEDIR=.
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')

LDFLAGS=-ldflags "-X main.GlobalVersion=${VERSION} -X main.BuildTime=${BUILD_TIME}"

.DEFAULT_GOAL: $(BINARY)

$(BINARY): $(SOURCES)
	go build ${LDFLAGS} -o ${BINARY} goxxx.go

.PHONY: install
install:
	go install ${LDFLAGS} ./...

.PHONY: clean
clean:
	if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi
