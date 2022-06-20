# protoc-gen-micro

This is protobuf code generation for micro. We use protoc-gen-micro to reduce boilerplate code.

## Install

```
go install github.com/pthethanh/micro/cmd/protoc-gen-micro
```

Also required: 

- [protoc](https://github.com/google/protobuf)
- [protoc-gen-go](https://google.golang.org/protobuf)

## Usage

1. Define your proto file normally, an example can be found [here](https://github.com/pthethanh/micro/blob/master/examples/helloworld/helloworld/helloworld.proto)
2. Generate code with option "micro_out", an example can be found [here](https://github.com/pthethanh/micro/blob/master/Makefile#L63)
3. Add option `--micro_opt generate_gateway=true` if you want to generate the gateway registration, see an example [here](https://github.com/pthethanh/micro/blob/master/Makefile#L63))
3. Register your service with micro server, an example can be found [here](https://github.com/pthethanh/micro/blob/master/examples/helloworld/server/main.go#L16) & [here](https://github.com/pthethanh/micro/blob/master/examples/helloworld/server/main.go#L37)

## LICENSE

protoc-gen-micro is a liberal reuse of protoc-gen-go hence we maintain the original license 