# micro

Just a simple tool kit for building microservices.

**Note**: Please notice that this is just in experiment stage right now and API can be changed without notification, hence please don't use this for production.

## What is micro?
micro is a Go tool kit for enterprise targeted for microservices or well designed monolith application. It doesn't aim to be a framework, but just a standard libraries for easily and quickly build API applications.

micro's vision is to be come a good tool kit for beginner/intermediate developers and hence it should be: 

- Easy to use.
- Compatible with Go native libraries.
- Come with default ready to use features. 

I expect micro requires no more than 15 minutes for a beginner/intermediate developer to be able to use the tool kit effectively. This means micro will come with lots of useful default features, but at the same time provide developers ability to provide their alternatives.

## Why a new standard libraries?

micro is inspired by [go-kit](<https://github.com/go-kit/kit>). go-kit is a good tool kit, but one of the thing I don't like go-kit is its over use of interface{} which cause a lot of unnecessary type conversions and some of other abstractions in the libraries which are not compatible with Go native libraries.

Although go-kit is very flexible, it's a little bit hard to use for beginner/intermediate developers. It has a lot of options for developers to choose and hence hard to force everyone inside a company to use the same set of standards.

Those reasons go against with my vision for micro, and this is the main reason I want to build micro.

## What in the early micro?

micro is now in the experiment stage and currently built around gRPC. It exposes both gRPC and REST API over 1 single port.

In this early experiment, I want to explore below things:

- gRPC
- Default, ready to use middleware:
  - Logging
  - Authentication
  - Authorization
  - Broker
  - Configuration
- Default, ready to use metrics for Prometheus
- API document: Swagger