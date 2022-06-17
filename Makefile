PROJECT_NAME=micro
GO_BUILD_ENV=CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on
GO_FILES=$(shell go list ./... | grep -v /vendor/)
PROTOC_VERSION = 3.10.1

GOPATH ?= $(HOME)/go
PROTO_OUT = $(GOPATH)/src
MOD=$(GOPATH)/pkg/mod
GOOGLE_APIS_PROTO := $(GOPATH)/src/github.com/googleapis/googleapis
PROTOC_INCLUDES := /usr/local/include

export PATH := $(GOPATH)/bin:$(PATH)

.SILENT:

all: mod_tidy fmt vet test build_plugins

stress: mod_tidy fmt vet test_stress build_plugins

micro: mod_tidy fmt vet test

vet:
	$(GO_BUILD_ENV) go vet $(GO_FILES)

fmt:
	$(GO_BUILD_ENV) go fmt $(GO_FILES)

test_stress:
	$(GO_BUILD_ENV) CGO_ENABLED=1 go test $(GO_FILES) -race -cover -count=25

test:
	$(GO_BUILD_ENV) CGO_ENABLED=1 go test $(GO_FILES) -race -cover -count=1

mod_tidy:
	$(GO_BUILD_ENV) go mod tidy
	$(GO_BUILD_ENV) go mod download

gen_proto: gen_proto_broker gen_proto_examples

install_tools:
	go install \
    github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
    github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
    google.golang.org/protobuf/cmd/protoc-gen-go \
    google.golang.org/grpc/cmd/protoc-gen-go-grpc

	git -C $(GOOGLE_APIS_PROTO) pull || git clone https://github.com/googleapis/googleapis $(GOOGLE_APIS_PROTO)

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
	$(PROTOC_ENV) protoc -I $(PROTOC_INCLUDES) -I $(GOOGLE_APIS_PROTO) -I ./broker/ --go_out $(PROTO_OUT) --go-grpc_out $(PROTO_OUT) broker/broker.proto

gen_proto_examples: install_tools
	$(PROTOC_ENV) protoc -I $(PROTOC_INCLUDES) -I $(GOOGLE_APIS_PROTO) -I ./examples/helloworld/helloworld \
	 --go_out $(PROTO_OUT) \
	 --go-grpc_out $(PROTO_OUT) \
	 --grpc-gateway_out $(PROTO_OUT) \
     --grpc-gateway_opt logtostderr=true \
     --grpc-gateway_opt generate_unbound_methods=true \
     ./examples/helloworld/helloworld/helloworld.proto

build_plugins:
	$(MAKE) -C  plugins/broker/nats
	$(MAKE) -C  plugins/cache/redis