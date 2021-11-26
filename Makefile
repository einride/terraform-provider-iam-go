SHELL := /bin/bash

HOSTNAME=hashicorp.com
NAMESPACE=einride
PKG_NAME=iamgo
NAME=iam-go
BINARY=terraform-provider-${NAME}
VERSION=0.1
OS_ARCH=linux_amd64

.PHONY: all
all: \
	go-lint \
	go-test \
	go-mod-tidy \
	git-verify-nodiff

include tools/git-verify-nodiff/rules.mk
include tools/golangci-lint/rules.mk
include tools/semantic-release/rules.mk

.PHONY: go-mod-tidy
go-mod-tidy:
	$(info [$@] tidying Go module files...)
	@go mod tidy

.PHONY: go-test
go-test:
	$(info [$@] running Go test suites...)
	go test -count=1 -race ./...

.PHONY: go-lint
go-lint: $(golangci_lint)
	$(info [$@] linting Go code...)
	@$(golangci_lint) run ./${PKG_NAME}

.PHONY: build
build:
	go build -o ${BINARY}

.PHONY: local-install
local-install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
