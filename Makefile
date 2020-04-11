PROJECT_NAME=micro
BUILD_VERSION=$(shell cat VERSION)
GO_BUILD_ENV=CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on
GO_FILES=$(shell go list ./... | grep -v /vendor/)

.SILENT:

all: mod_tidy fmt vet build test

vet:
	$(GO_BUILD_ENV) go vet $(GO_FILES)

fmt:
	$(GO_BUILD_ENV) go fmt $(GO_FILES)

test:
	$(GO_BUILD_ENV) go test $(GO_FILES) -cover -v

mod_tidy:
	$(GO_BUILD_ENV) go mod tidy

build:
	$(GO_BUILD_ENV) go build -v  $(GO_FILES)