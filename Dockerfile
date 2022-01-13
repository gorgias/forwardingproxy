# Copyright (C) 2018 Betalo AB - All Rights Reserved

FROM golang:1.10

LABEL maintainer="Betalo Backend Team <backend-team@betalo.se>"

RUN mkdir -p /go/src/github.com/gorgias/forwardingproxy
WORKDIR /go/src/github.com/gorgias/forwardingproxy

COPY ./vendor vendor
COPY *.go ./

RUN go install

EXPOSE 80 443

ENTRYPOINT ["forwardingproxy"]
