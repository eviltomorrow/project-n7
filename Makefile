# This how we want to name the binary output
#
MAINVERSION=$(shell cat version)
GOPATH ?= $(shell go env GOPATH)
ifeq "$(GOPATH)" ""
  $(error Please set the environment variable GOPATH before running `make`)
endif
PATH := ${GOPATH}/bin:$(PATH)
GCFLAGS=-gcflags "all=-trimpath=${GOPATH}"
GITSHA := $(shell git rev-parse HEAD)
BUILDTIME=`date +%FT%T%z`
LDFLAGS=-ldflags "-X main.MainVersion=${MAINVERSION} -X main.GitSha=${GITSHA} -X main.BuildTime=${BUILDTIME} -s -w"

# colors compatible setting
CRED:=$(shell tput setaf 1 2>/dev/null)
CGREEN:=$(shell tput setaf 2 2>/dev/null)
CYELLOW:=$(shell tput setaf 3 2>/dev/null)
CEND:=$(shell tput sgr0 2>/dev/null)

.PHONY: go_version_check
GO_VERSION_MIN=1.19
# Parse out the x.y or x.y.z version and output a single value x*10000+y*100+z (e.g., 1.9 is 10900)
# that allows the three components to be checked in a single comparison.
VER_TO_INT:=awk '{split(substr($$0, match ($$0, /[0-9\.]+/)), a, "."); print a[1]*10000+a[2]*100+a[3]}'
go_version_check:
	@echo "$(CGREEN)=> Go version check ...$(CEND)"
	@if test $(shell go version | $(VER_TO_INT) ) -lt \
  	$(shell echo "$(GO_VERSION_MIN)" | $(VER_TO_INT)); \
  	then printf "go version $(GO_VERSION_MIN)+ required, found: "; go version; exit 1; \
		else echo "go version check pass";	fi

# Code format
.PHONY: fmt
fmt: go_version_check
	@echo "$(CGREEN)=> Run gofmt on all source files ...$(CEND)"
	@echo "gofmt -l -s -w ..."
	@ret=0 && for d in $$(go list -f '{{.Dir}}' ./... | grep -v /vendor/); do \
		gofmt -l -s -w $$d/*.go || ret=$$? ; \
	done ; exit $$ret


# Compile protobuf
.PHONY: compile	
compile:
	@echo "$(CGREEN)=> Compile protobuf ...$(CEND)"
	@bash build/protobuf_compile.sh

# build
.PHONY: build
build: export CGO_ENABLED=0
build: 
ifeq (${app},)
	@bash build/app_build.sh ${LDFLAGS} ${GCFLAGS}
else
	@rm -rf bin/${app}/*
	@mkdir -p bin/${app}/etc
	@cp -rp app/${app}/etc bin/${app}/
	@echo "$(CGREEN)=> Building binary(${app})...$(CEND)"
	go build ${LDFLAGS} ${GCFLAGS} -o bin/${app}/bin/${app} app/${app}/main.go
	@echo "$(CGREEN)=> Build Success!$(CEND)"
endif

# build
.PHONY: docker
docker: 
ifeq (${app},)
	@bash build/docker_build.sh ${MAINVERSION} ${GITSHA} ${BUILDTIME}
else
	docker build --target prod -t ${app} . --build-arg APPNAME=${app} --build-arg MAINVERSION=${MAINVERSION} --build-arg GITSHA=${GITSHA} --build-arg BUILDTIME=${BUILDTIME}
endif

# clear
.PHONY: clear
clear:
	@echo "$(CGREEN)=> Clear ./bin...$(CEND)"
	@rm -rf bin/n7-*

# Go mod tidy
.PHONY: mod
mod:export GO111MODULE=on
mod:
	@echo "$(CGREEN)=> go mod tidy...$(CEND)"
	@go mod tidy

# Go mod vendor
.PHONY: vendor
vendor:export GO111MODULE=on
vendor:
	@echo "$(CGREEN)=> go mod vendor...$(CEND)"
	@go mod vendor
