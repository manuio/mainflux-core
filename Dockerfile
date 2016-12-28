###
# Mainflux Dockerfile
###

FROM golang:alpine
MAINTAINER Mainflux

ENV MONGO_HOST mongo
ENV MONGO_PORT 27017

ENV NATS_HOST nats
ENV NATS_PORT 4222

###
# Install
###

RUN apk update && apk add git && apk add wget && rm -rf /var/cache/apk/*

# Copy the local package files to the container's workspace.
ADD . /go/src/github.com/mainflux/mainflux-core

RUN mkdir -p /etc/mainflux/core
COPY config/config-docker.toml /etc/mainflux/core/config.toml

# Get and install the dependencies
RUN go get github.com/mainflux/mainflux-core

# Dockerize
ENV DOCKERIZE_VERSION v0.2.0
RUN wget https://github.com/jwilder/dockerize/releases/download/$DOCKERIZE_VERSION/dockerize-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
	&& tar -C /usr/local/bin -xzvf dockerize-linux-amd64-$DOCKERIZE_VERSION.tar.gz

###
# Run main command with dockerize
###
CMD dockerize -wait tcp://$MONGO_HOST:$MONGO_PORT \
				-wait tcp://$NATS_HOST:$NATS_PORT \
				-timeout 10s /go/bin/mainflux-core /etc/mainflux/core/config.toml

