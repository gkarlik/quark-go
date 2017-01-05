FROM golang

MAINTAINER Carlos

LABEL service="discovery"

COPY . /go/src/carlos/micro

WORKDIR /go/src/carlos/micro

# install govendor
RUN go get -u github.com/kardianos/govendor

# synchronize dependencies
RUN govendor sync

# install discovery service
RUN go install carlos/micro/services/discovery

# move service configuration to bin directory
RUN mv services/discovery/discovery_service.yaml /go/bin/discovery_service.yaml

WORKDIR /go/bin

# remove sources
RUN rm -r /go/src
RUN rm /go/bin/govendor
RUN rm -r /go/pkg

ENTRYPOINT discovery -configuration="discovery_service.yaml"

EXPOSE 32000:32000