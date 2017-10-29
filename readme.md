# quark-go [![Go Report Card](https://goreportcard.com/badge/github.com/gkarlik/quark-go)](https://goreportcard.com/report/github.com/gkarlik/quark-go) [![Build Status](https://travis-ci.org/gkarlik/quark-go.svg?branch=master)](https://travis-ci.org/gkarlik/quark-go) [![Coverage Status](https://coveralls.io/repos/github/gkarlik/quark-go/badge.svg?branch=master)](https://coveralls.io/github/gkarlik/quark-go?branch=master) [![GoDoc](https://godoc.org/github.com/gkarlik/quark-go?status.svg)](https://godoc.org/github.com/gkarlik/quark-go)


Quark-go is a quark size (meaning very very small) toolbelt for building microservices in golang. 

**Important:** Library requires Golang 1.8+! Work in progress! Some interfaces could be changed!

## Goals
The goal of the project is to help quickly build microservices using distributed programming best practices and tools which are 
best in the class (choice is subjective). Common techniques and components are at disposal of a developer who should be 
focus more on business logic instead of tweaking and finding right tools to do the job.

Quark-go is very extensible and allows to implement custom providers for all specified features below. It is not the goal of the project
to support all available tools, configurations and components on the market. Project is focused to deliver community proven best preconfigured tools
and components prepared OOTB to use it in your projects. It aims to be end-to-end solution for modern, distributed applications.

If you are interesed in more "enterprise" solutions. Please see the following projects:
* [go-kit](https://github.com/go-kit/kit)
* [go-micro](https://github.com/micro/go-micro)

## Features
* **Message Broker** - asynchronous messaging using [RabbitMQ](https://www.rabbitmq.com/) and [Apache Kafka](https://kafka.apache.org/)
* **Circuit Breaker** - custom implementation of [Circuit Breaker pattern](https://martinfowler.com/bliki/CircuitBreaker.html)
* **Logging** - structured service diagnostics using [Logrus](https://github.com/sirupsen/logrus) library
* **Metrics Collection** - service metrics collection using [Prometheus](https://prometheus.io/)
* **Service Discovery** - service discovery using [Consul](https://www.consul.io/)
* **Load Balancing** - custom implementation of load balancing strategy
* **RPC** - Remote Procedure Call client and server using [gRPC](http://www.grpc.io/) library
* **Request Tracing** - using [opentracing](http://opentracing.io/) and [zipkin](http://zipkin.io/)
* **Data Access Layer**
    * **Relational databases** - using Domain-Driven Design [aggregates](https://martinfowler.com/bliki/DDD_Aggregate.html) and [repository](https://martinfowler.com/eaaCatalog/repository.html) pattern 
* **Middlewares:**
    * **Authentication** - middleware for HTTP [JSON Web Tokens](https://jwt.io/) authentication
    * **Rate Limiter** - custom implementation of HTTP rate limiter middleware
    * **Error** - panic error recovery middleware
    * **Logging** - request logging middleware
    * **Metrics** - middleware for request metrics exposing
    * **Security** - middleware for securing request via HTTP headers
    * **Tracing** - request tracing middleware

## Planned features
* **More security** - HTTP headers, OpenID Connect, Two-factor authentication etc.
* **Searchability** - Elasticsearch indexing and searching
* **Data Access Layer** - patterns for accessing data (document oriented)
* **Caching** - data caching patterns

## Installation

`$ go get -u github.com/gkarlik/quark-go`

## Examples

Please see repo with [example](https://github.com/gkarlik/quark-go-example) and [documentation](https://godoc.org/github.com/gkarlik/quark-go).