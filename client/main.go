package main

import (
	"time"
	"flag"
	"fmt"
	"github.com/google/uuid"
	"log"
	"os"
)

var serverAddr = os.Getenv("SERVER_ADDR")

func main() {
	clientsNum := flag.Int("c", 1, "Number of clients")
	requestsNum := flag.Int("r", 1000, "Number of requests")
	flag.Parse()

	startTime := time.Now()
	done := make(chan struct{})
	for i := 0; i < *clientsNum; i++ {
		clientId := int64(uuid.New().ID())
		ag := NewActivityGenerator(clientId, fmt.Sprintf("http://%v/api/activity/v1/", serverAddr), done, log.New(os.Stdout, "activity:", 0))
		vg := NewVisitGenerator(clientId, fmt.Sprintf("http://%v/api/visit/v1/", serverAddr), done, log.New(os.Stdout, "visit:", 0))
		go vg.Run(*requestsNum)
		go ag.Run(*requestsNum)
	}

	for i := 0; i < *clientsNum; i++ {
		<- done // visit
		<- done // activity
	}

	fmt.Println("Done in:", time.Now().Sub(startTime).Seconds(), "seconds")
	fmt.Printf("Sent %v requests total\n", int(*clientsNum)*(*requestsNum)*2)
}



