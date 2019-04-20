FROM golang:latest

RUN go get -u github.com/segmentio/kafka-go
RUN go get -u github.com/google/uuid
RUN go get -u github.com/olesho/kfk/client

WORKDIR /go/src/github.com/olesho/kfk/client
ADD . .

RUN go build .
CMD ./client  -c 1 -r 100