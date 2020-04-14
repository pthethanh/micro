# micro
[![Actions Status](https://github.com/pthethanh/micro/workflows/Go/badge.svg)](https://github.com/pthethanh/micro/actions)
[![GoDoc](https://godoc.org/github.com/pthethanh/micro?status.svg)](https://pkg.go.dev/mod/github.com/pthethanh/micro)
[![GoReportCard](https://goreportcard.com/badge/github.com/pthethanh/micro)](https://goreportcard.com/report/github.com/pthethanh/micro)


Just a simple tool kit for building microservices.

**Note**: Please notice that this is just in experiment stage right now and API can be changed without notification, hence please don't use this for production.

## What is micro?
micro is a Go tool kit for enterprise targeted for microservices or well designed monolith application. It doesn't aim to be a framework, but just a standard libraries for easily and quickly build API applications.

micro's vision is to be come a good tool kit for beginner/intermediate developers and hence it should be: 

- Easy to use.
- Compatible with Go native libraries.
- Come with default ready to use features. 

I expect micro requires no more than 15 minutes for a beginner/intermediate developer to be able to use the tool kit effectively. This means micro will come with lots of useful default features, but at the same time provide developers ability to provide their alternatives.

micro is now in the experiment stage and currently built around gRPC. It exposes both gRPC and REST API over 1 single port using [grpc-gateway](https://github.com/grpc-ecosystem/grpc-gateway).

Currently, micro supports following components:

- Default, ready to use server:
  - Exposes both gRPC and REST in 1 single port.
  - Internal APIs: metrics, health checks
  - Context logger with request-id
- Default, ready to use middleware:
  - Logging
  - Authentication
  - Authorization
- Default, ready to use:
  - Broker
  - Configuration
  - API document: Swagger

## Why a new standard libraries?

micro is inspired by [go-kit](https://github.com/go-kit/kit) and [go-micro](https://github.com/micro/go-micro). 

go-kit is a good tool kit, but one of the thing I don't like go-kit is its over use of interface{} which cause a lot of unnecessary type conversions and some of other abstractions in the libraries which are not compatible with Go native libraries. Although go-kit is very flexible, it's a little bit hard to use for beginner/intermediate developers. It has a lot of options for developers to choose and hence hard to force everyone inside a company to use the same set of standards.

go-micro is a great framework for microservices and very well designed. And it influences micro very much, but there are some coding styles that I don't like go-micro, that's why I made micro for myself.


## Documentation

- See [doc](https://pkg.go.dev/mod/github.com/pthethanh/micro) for package and API descriptions.
- Examples can be found in the [examples](https://github.com/pthethanh/micro/tree/master/examples) directory.