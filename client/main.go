package main

import (
	"time"
	"flag"
	"fmt"
	"github.com/google/uuid"
	"log"
	"os"
)

func main() {
	clientsNum := flag.Int("c", 1, "Number of clients")
	host := flag.String("h", "localhost", "Remote host name")
	requestsNum := flag.Int64("r", 1000, "Number of requests")
	flag.Parse()

	startTime := time.Now()
	done := make(chan struct{})
	for i := 0; i < *clientsNum; i++ {
		clientId := int64(uuid.New().ID())
		ag := NewActivityGenerator(clientId, "http://" + *host + ":8000/api/activity/v1/", done, log.New(os.Stdout, "activity:", 0))
		vg := NewVisitGenerator(clientId, "http://" + *host + ":8000/api/visit/v1/", done, log.New(os.Stdout, "activity:", 0))
		go vg.Run(*requestsNum)
		go ag.Run(*requestsNum)
	}

	for i := 0; i < *clientsNum; i++ {
		<- done // visit
		<- done // activity
	}

	fmt.Println("Done in:", time.Now().Sub(startTime).Seconds(), "seconds")
	fmt.Printf("Sent %v requests total\n", int64(*clientsNum)*(*requestsNum)*2)
}



