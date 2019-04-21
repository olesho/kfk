package main

import (
	"time"
	"fmt"
	"github.com/google/uuid"
	"log"
	"os"
	"net/http"
	"net/url"
	"strconv"
)

var serverAddr = os.Getenv("SERVER_ADDR")
var port = os.Getenv("PORT")

func main() {
	threadsNum := 1
	requestsNum := 1000

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Client (requests generator) service example"))
	})

	http.HandleFunc("/generate/", func(w http.ResponseWriter, r *http.Request) {
		m, _ := url.ParseQuery(r.URL.RawQuery)
		if arr, ok := m["threads"]; ok {
			if len(arr) > 0 {
				threadsNum, _ = strconv.Atoi(arr[0])
			}
		}
		if arr, ok := m["requests"]; ok {
			if len(arr) > 0 {
				requestsNum, _ = strconv.Atoi(arr[0])
			}
		}

		startTime := time.Now()
		generate(threadsNum, requestsNum)
		w.Write([]byte(fmt.Sprintf("Done in:", time.Now().Sub(startTime).Seconds(), "seconds\n")))
		w.Write([]byte(fmt.Sprintf("Sent %v requests total\n", threadsNum*requestsNum*2)))

		log.Printf("Done in %v seconds\n", time.Now().Sub(startTime).Seconds())
		log.Printf("Sent %v requests total\n", threadsNum*requestsNum*2)
	})

	if err := http.ListenAndServe(fmt.Sprintf(":%v", port), nil); err != nil {
		panic(err)
	}
}

func generate(threads, requestsPerThread int) {
	vgs := make([]*VisitGenerator, threads)
	ags := make([]*ActivityGenerator, threads)

	for i := 0; i < threads; i++ {
		clientId := int64(uuid.New().ID())
		ags[i] = NewActivityGenerator(clientId, fmt.Sprintf("http://%v/api/activity/v1/", serverAddr), log.New(os.Stdout, "activity:", 0))
		vgs[i] = NewVisitGenerator(clientId, fmt.Sprintf("http://%v/api/visit/v1/", serverAddr), log.New(os.Stdout, "visit:", 0))
		go vgs[i].Run(requestsPerThread)
		go ags[i].Run(requestsPerThread)
	}

	for i := 0; i < threads; i++ {
		<- ags[i].DoneSignal() // wait for visit generation finish
		<- vgs[i].DoneSignal() // wait for activity generation finish
	}
}

