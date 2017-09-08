# It's necessary to set this because some environments don't link sh -> bash.
SHELL := /bin/bash

SOURCES := $(shell git ls-files '*.go' | grep -v '^vendor/')
SOURCE_PKGS := $(shell go list ./... | grep -v 'vendor/')
OS_TARGETS ?= linux,darwin,windows
ARCH_TARGETS ?= 386,amd64

APP_NAME := gbl
APP_BIN := $(APP_NAME)_$(shell go env GOOS)_$(shell go env GOARCH)

IS_TAGGED := $(TRAVIS_TAG)

default: prepare build

.PHONY: help
help:
	@echo "help:        		Show usage."
	@echo "run:         		Run application."
	@echo "prepare:     		Install extra build tools."
	@echo "build:       		Build code."
	@echo "test:        		Run tests."
	@echo "clean:       		Clean up."

.PHONY: run
run:
	go run bilibili.go

.PHONY: debug-run
debug-run:
	GODEBUG=gctrace=1,schedtrace=10000,scheddetail=1 go run bilibili.go

.PHONY: binary
binary: clean vet lint test
	go build -x -v -i -o $(APP_BIN)
	file $(APP_BIN)
	@echo "Build code success."

.PHONY: install
install:
	cp ./$(APP_BIN) $(GOPATH)/bin
	@echo "Install success."

.PHONY: uninstall
uninstall: clean
	-@rm -f $(GOPATH)/bin/$(APP_BIN)
	@echo "Uninstall success."

.PHONY: release
release: clean vet lint test
	CGO_ENABLED=0 goxc -arch="$(ARCH_TARGETS)" -bc="$(OS_TARGETS)" -v -build-verbose="true" \
	  -tasks-="validate,archive,deb,deb-dev,rmbin,downloads-page" -build-print-commands="true" \
	  -build-ldflags="-s -w" -o="gbl_{{.Os}}_{{.Arch}}{{.Ext}}"
	@echo "Release success."

.PHONY: build
ifdef IS_TAGGED
build: release
else
build: binary
endif

.PHONY: prepare
prepare:
	@hash goxc > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		go get -u -v github.com/laher/goxc; \
	fi
	@hash golint > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		go get -u -v github.com/golang/lint/golint; \
	fi

.PHONY: update-tools
update-tools:
	go get -u -v github.com/laher/goxc
	go get -u -v github.com/golang/lint/golint

.PHONY: vet
vet:
	@echo "vet ..."
	@for gocode in $(SOURCES); \
	do \
		if [ -f $$gocode ]; then \
			go tool vet -all -shadow -shadowstrict $$gocode; \
		fi \
	done;
	@echo "vet done"

.PHONY: lint
lint:
	@echo "lint ..."
	-@if [ "`golint $(SOURCE_PKGS) | tee /dev/stderr`" ]; then \
		echo "^ golint errors!" && echo && exit 1; \
	fi
	@echo "lint done"

.PHONY: test
test:
	go test -v ./...

.PHONY:
clean:
	-@rm -f ./gbl*
	-@rm -f ./go-bilibili-live*
	-@rm -f *.out
	go clean -i ./...
