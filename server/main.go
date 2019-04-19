package main

import (
	"os"
)

var kafkaAddr = os.Getenv("KAFKA_ADDR")

func main() {
	serve()
}