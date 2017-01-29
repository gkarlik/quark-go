FROM golang

RUN go get github.com/dgrijalva/jwt-go
RUN go get github.com/stretchr/testify/assert
RUN go get golang.org/x/net/context
RUN go get github.com/streadway/amqp
RUN go get github.com/Sirupsen/logrus
RUN go get github.com/influxdata/influxdb/client/v2
RUN go get golang.org/x/time/rate
RUN go get github.com/hashicorp/consul/api
RUN go get google.golang.org/grpc
RUN go get github.com/opentracing/opentracing-go
RUN go get github.com/openzipkin/zipkin-go-opentracing

COPY . /go/src/github.com/gkarlik/quark-go

ENTRYPOINT /bin/bash
