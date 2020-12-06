PROJECT_NAME=micro
BUILD_VERSION=$(shell cat VERSION)
GO_BUILD_ENV=CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on
GO_FILES=$(shell go list ./... | grep -v /vendor/)

GRPC_GATEWAY_VERSION = $(shell go list -m -f '{{ .Version }}' github.com/grpc-ecosystem/grpc-gateway/v2)
PROTOC_VERSION = 3.10.1

GOPATH ?= $(HOME)/go
PROTO_OUT = $(GOPATH)/src
MOD=$(GOPATH)/pkg/mod
GRPC_GATEWAY_INCLUDES := $(MOD)/github.com/grpc-ecosystem/grpc-gateway/v2@$(GRPC_GATEWAY_VERSION)/third_party/googleapis
PROTOC_INCLUDES := /usr/local/include
PROTOC_GEN_GO = $(GOPATH)/bin/protoc-gen-go
PROTOC_GEN_GRPC_GATEWAY = $(GOPATH)/bin/protoc-gen-grpc-gateway


.SILENT:

all: fmt vet build test

vet:
	$(GO_BUILD_ENV) go vet $(GO_FILES)

fmt:
	$(GO_BUILD_ENV) go fmt $(GO_FILES)

test:
	$(GO_BUILD_ENV) go test $(GO_FILES) -cover -count=1

mod_tidy:
	$(GO_BUILD_ENV) go mod tidy

build:
	$(GO_BUILD_ENV) go build -v  $(GO_FILES)

gen_proto: gen_proto_broker gen_proto_examples

install_tools:
	go install \
    github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
    github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
    google.golang.org/protobuf/cmd/protoc-gen-go \
    google.golang.org/grpc/cmd/protoc-gen-go-grpc

install_protobuf:
	wget https://github.com/google/protobuf/releases/download/v$(PROTOC_VERSION)/protoc-$(PROTOC_VERSION)-linux-x86_64.zip
	unzip protoc-$(PROTOC_VERSION)-linux-x86_64.zip -d protoc
	sudo cp protoc/bin/protoc /usr/local/bin
	sudo mkdir -p /usr/local/include
	sudo cp -R protoc/include/* /usr/local/include/
	rm -rf protoc
	rm -rf protoc-$(PROTOC_VERSION)-linux-x86_64.zip
	sudo chmod -R 755 /usr/local/include/
	sudo chmod +x /usr/local/bin/protoc

gen_proto_broker: install_tools
	protoc -I $(PROTOC_INCLUDES) -I $(GRPC_GATEWAY_INCLUDES) -I ./broker/ --go_out $(PROTO_OUT) --go-grpc_out $(PROTO_OUT) broker/broker.proto

gen_proto_examples: install_tools
	protoc -I $(PROTOC_INCLUDES) -I $(GRPC_GATEWAY_INCLUDES) -I ./examples/helloworld/helloworld \
	 --go_out $(PROTO_OUT) \
	 --go-grpc_out $(PROTO_OUT) \
	 --grpc-gateway_out $(PROTO_OUT) \
     --grpc-gateway_opt logtostderr=true \
     --grpc-gateway_opt generate_unbound_methods=true \
     ./examples/helloworld/helloworld/helloworld.proto