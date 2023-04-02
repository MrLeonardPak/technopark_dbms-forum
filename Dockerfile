FROM golang:alpine
COPY . /go/src/api

RUN wget https://github.com/prometheus/node_exporter/releases/download/v1.5.0/node_exporter-1.5.0.linux-amd64.tar.gz \
    && tar xvfz node_exporter-*.*-amd64.tar.gz \
    && rm -rf node_exporter-*.*-amd64.tar.gz

RUN cd src/api \
    && CGO_ENABLED=0 go build -o /go/bin/api main.go

ENTRYPOINT /go/node_exporter-*.*-amd64/node_exporter & /go/bin/api
