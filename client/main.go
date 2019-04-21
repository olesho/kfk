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

	http.HandleFunc("/generate", func(w http.ResponseWriter, r *http.Request) {
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
		generate(threadsNum, requestsNum)
	})

	if err := http.ListenAndServe(fmt.Sprintf(":%v", port), nil); err != nil {
		panic(err)
	}
}

func generate(threads, requestsPerThread int) {
	vgs := make([]*VisitGenerator, threads)
	ags := make([]*ActivityGenerator, threads)

	startTime := time.Now()
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

	fmt.Println("Done in:", time.Now().Sub(startTime).Seconds(), "seconds")
	fmt.Printf("Sent %v requests total\n", threads*requestsPerThread*2)
}

