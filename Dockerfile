FROM golang:alpine
COPY . /go/src/api

RUN cd src/api \
    && CGO_ENABLED=0 go build -o /go/bin/api main.go

ENTRYPOINT /go/bin/api
