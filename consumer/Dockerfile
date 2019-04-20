FROM golang:latest

RUN go get -u github.com/segmentio/kafka-go
RUN go get -u github.com/golang/snappy
RUN go get -u github.com/olesho/kfk/structs
RUN go get -u github.com/olesho/kfk/consumer

WORKDIR /go/src/github.com/olesho/kfk/consumer
ADD . .

RUN go build . 
#CMD ./consumer

CMD tail -f /dev/null