package main

import (
	"net/http"
	"github.com/olesho/kfk/structs"
	"encoding/json"
	"os"
	"io"
	"compress/gzip"
	"log"
	"bufio"
	"github.com/segmentio/kafka-go"
	"time"
	"context"
	"strings"
	"strconv"
	"fmt"
	"math/rand"
)

var DeleteTopicHandler = func(w http.ResponseWriter, r *http.Request) {
	err := DeleteTopic("visit")
	if err != nil {
		fmt.Println(err)
	}
	err = DeleteTopic("activity")
	if err != nil {
		fmt.Println(err)
	}
}

var ListTopicsHandler = func(w http.ResponseWriter, r *http.Request) {
	ids, err := ListPartitions("visit")
	if err != nil {
		log.Println(err)
	} else {
		fmt.Println("'visit' partitions:", ids)
	}

	ids, err = ListPartitions("activity")
	if err != nil {
		log.Println(err)
	} else {
		fmt.Println("'activity' partitions:", ids)
	}
}

var CreateTopicHandler = func(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.RequestURI, "/")
	numPartitions, _ := strconv.Atoi(parts[len(parts)-1])

	err := CreateTopic("visit", numPartitions)
	if err != nil {
		log.Println(err)
	}

	err = CreateTopic("activity", numPartitions)
	if err != nil {
		log.Println(err)
	}
}

var LoadHandler = func(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	dialer := &kafka.Dialer{
		Timeout:  10 * time.Second,
		ClientID: "1",
	}

	config := kafka.WriterConfig{
		BatchSize:		  1,
		Brokers:          []string{kafkaAddr},
		Topic:            "visit",
		//Balancer:			&kafka.Hash{},
		//Balancer:         &kafka.LeastBytes{},
		Dialer:           dialer,
		WriteTimeout:     10 * time.Second,
		ReadTimeout:      10 * time.Second,
		//CompressionCodec: snappy.NewCompressionCodec(),
	}
	visitWriter := kafka.NewWriter(config)

	defer visitWriter.Close()

	messages := []kafka.Message{}

	sent := 0
	startTime, _ := time.Parse("2006-01-02T15:04:05", "2018-01-01T00:00:00")
	for i := 0; i < 1666; i++ {
		v := &structs.VisitPayload{
			DataVer:       1,
			UserId:        0,
			EnterTime:    	startTime.UnixNano(),
			ExitTime:      	startTime.Add(time.Minute*40).UnixNano(),
			AlgorithmType: rand.Intn(6)+1,
			PoiId:         rand.Int63(),
			Latitude:      rand.Int63(),
			Longitude:     rand.Int63(),
		}

		messages = append(messages, kafka.Message{
			Value: 	v.Bytes(),
			Time:	time.Now(),
		})
		startTime = startTime.Add(time.Hour*2)

		if (len(messages) > 1665) || (sent+len(messages)) == 1666 {
			err := visitWriter.WriteMessages(context.Background(), messages...)
			if err != nil {
				panic(err)
			} else {
				sent += len(messages)
				fmt.Println("sent", sent)
			}
			messages = []kafka.Message{}
		}
	}

	fmt.Println("Done in:", time.Now().Sub(start).Seconds())
}

var VisitHandler = func(w http.ResponseWriter, r *http.Request) {
	var reader io.ReadCloser
	switch r.Header.Get("Content-Encoding") {
	case "gzip":
		var err error
		reader, err = gzip.NewReader(r.Body)
		if err != nil {
			log.Println(err)
		}
		defer reader.Close()
	default:
		reader = r.Body
	}

	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		v := &structs.VisitPayload{}
		err := json.Unmarshal(scanner.Bytes(), v)
		if err != nil {
			log.Println(err)
		}

		// 1 MSG
		messages <- kafka.Message{
			Value: v.Bytes(),
			//Time:  time.Now(),
		}
	}

	if err := scanner.Err(); err != nil {
		log.Println(err)
	}
}

var ActivityHandler = func(w http.ResponseWriter, r *http.Request) {
	var reader io.ReadCloser
	switch r.Header.Get("Content-Encoding") {
	case "gzip":
		var err error
		reader, err = gzip.NewReader(r.Body)
		if err != nil {
			log.Println(err)
		}
		defer reader.Close()
	default:
		reader = r.Body
	}

	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		v := &structs.ActivityPayload{}
		err := json.Unmarshal(scanner.Bytes(), v)
		if err != nil {
			log.Println(err)
		}

		// 1 MSG
		messages <- kafka.Message{
			Value: v.Bytes(),
			//Time:  time.Now(),
		}
	}

	if err := scanner.Err(); err != nil {
		log.Println(err)
	}
}

var ShutdownHandler = func(w http.ResponseWriter, r *http.Request) {
	os.Exit(0)
}

var messages = make(chan kafka.Message, 500)

func serve() {
	visitWriter, err := ConfigureBatchWriter([]string{kafkaAddr}, "1", "visit")
	if err != nil {
		log.Println(err)
		return
	}
	defer visitWriter.Close()

	activityWriter, err := ConfigureBatchWriter([]string{kafkaAddr}, "1", "activity")
	if err != nil {
		log.Println(err)
		return
	}
	defer activityWriter.Close()

	runStream(visitWriter)
	runStream(activityWriter)


	http.HandleFunc("/api/topic/load/v1/", LoadHandler) // test
	http.HandleFunc("/api/topic/delete/v1/", DeleteTopicHandler) // delete topics
	http.HandleFunc("/api/topic/list/v1/", ListTopicsHandler) // list topics
	http.HandleFunc("/api/topic/create/v1/", CreateTopicHandler) // create topics
	http.HandleFunc("/api/shutdown/v1/", ShutdownHandler) // exit app
	http.HandleFunc("/api/visit/v1/", VisitHandler)
	http.HandleFunc("/api/activity/v1/", ActivityHandler)

	http.ListenAndServe(":8000", nil)
}

func send(writer *kafka.Writer, slice []kafka.Message, timer *time.Timer) {
	err := writer.WriteMessages(context.Background(), slice...)
	if err != nil {
		log.Println(err)
	}
	slice = []kafka.Message{}
	timer.Reset(time.Second)
}


func runStream(writer *kafka.Writer) *kafka.Writer  {
	slice := []kafka.Message{}
	timer := time.NewTimer(time.Second)

	go func () {
		for {
			slice = append(slice, <- messages)
			if len(slice) == 500 {
				send(writer, slice, timer)
			}
		}
	}()

	go func() {
		for {
			<- timer.C
			send(writer, slice, timer)
		}
	}()
	return writer
}