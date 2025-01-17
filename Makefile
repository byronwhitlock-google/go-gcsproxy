OUT := bin/go-gcsproxy
PKG := github.com/byronwhitlock-google/go-gcsproxy
VERSION := $(shell git describe --always --long --dirty)
#PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/ | grep -v /test/)
GO_FILES := $(shell find . -name '*.go' | grep -v /vendor/ | grep -v /test/)

$(info Go binary location: $(shell which go))

all: run

server:
	go build -o ${OUT} -ldflags="-X main.version=${VERSION}" ${PKG}

test:
	@go test -short ${GO_FILES}

vet:
	@go vet ${GO_FILES}

lint:
	@for file in ${GO_FILES} ;  do \
		golint $$file ; \
	done

static: vet lint
A	go build -i -v -o ${OUT}-v${VERSION} -tags netgo -ldflags="-extldflags \"-static\" -w -s -X main.version=${VERSION}" ${PKG}

run: server
	./${OUT}

clean:
	-@rm ${OUT} ${OUT}-v*

.PHONY: run server static vet lint