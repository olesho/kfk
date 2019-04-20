FROM golang:latest

RUN go get -u github.com/segmentio/kafka-go
RUN go get -u github.com/olesho/kfk/structs
RUN go get -u github.com/olesho/kfk/server

WORKDIR /go/src/github.com/olesho/kfk/server
ADD . .

RUN go build .
CMD ./server

#CMD tail -f /dev/null