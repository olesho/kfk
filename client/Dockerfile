FROM golang:latest
ADD . .
RUN go get ./...
RUN go build -o main . 
CMD ["./main"]
