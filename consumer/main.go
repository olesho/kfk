package main

import (
	"os"
	"github.com/segmentio/kafka-go"
	"context"
	"time"
	"fmt"
	_ "github.com/segmentio/kafka-go/snappy"
	"github.com/olesho/kfk/structs"
	"encoding/json"
	"sync/atomic"
	"log"
	"net/http"
)

func Reader(kafkaBrokerUrls []string, kafkaTopic, groupId string, partition int) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Partition:		partition,
		Brokers:		kafkaBrokerUrls,
		GroupID:        groupId,
		Topic:			kafkaTopic,
		MinBytes:		10e3,            // 10KB
		MaxBytes:		10e6,            // 10MB
		MaxWait:		1 * time.Second, // Maximum amount of time to wait for new data to come when fetching batches of messages from kafka.
		ReadLagInterval:	-1,
	})
	for {
		readBatch(reader, partition, kafkaTopic)
	}
}

func readBatch(reader *kafka.Reader, partition int, kafkaTopic string) {
	var cnt = new(uint32)
	var encoder *json.Encoder

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*2)
	defer cancel()

	fileName := fmt.Sprintf("%v/%v_%v_%v.json", storageDir, kafkaTopic, partition, time.Now().Format("2006-01-02_15-04-05.999"))
	f, err := os.Create(fileName)
	if err != nil {
		log.Println(err)
		return
	}
	defer f.Close()


	encoder = json.NewEncoder(f)
	*cnt = 0

	for *cnt < 100 {
		m, err := reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				log.Println("deadline exceeded")
			} else {
				log.Println("error while receiving message: ", err.Error())
			}
			return
		} else {
			if kafkaTopic == "visit" {
				p, err := structs.VisitPayloadFromBytes(m.Value)
				if err != nil {
					log.Println(err)
				} else {
					err = encoder.Encode(p)
					if err != nil {
						log.Println("error while encoding message: ", err.Error())
					} else {
						atomic.AddUint32(cnt, 1)
					}
				}
			} else if kafkaTopic == "activity" {
				p, err := structs.ActivityPayloadFromBytes(m.Value)
				if err != nil {
					log.Println(err)
				} else {
					err = encoder.Encode(p)
					if err != nil {
						log.Println("error while encoding message: ", err.Error())
					} else {
						atomic.AddUint32(cnt, 1)
					}
				}
			}

			err = reader.CommitMessages(context.Background(), m)
			if err != nil {
				log.Println(err)
			}
		}
	}

	atomic.AddUint32(globalCnt, *cnt)
}

var globalCnt = new(uint32)
var storageDir = "./files"

func main() {
	if _, err := os.Stat("files"); os.IsNotExist(err) {
		os.Mkdir("files", os.ModePerm)
	}

	var kafkaAddr = os.Getenv("KAFKA_ADDR")
	var kafkaTopic = os.Getenv("TOPIC")
	var groupId = os.Getenv("GROUP_ID")
	var port = os.Getenv("PORT")

	var partitions []kafka.Partition

	// try dialing until Kafka is up and connection established
	var conn *kafka.Conn
	for {
		var err error
		time.Sleep(time.Second*1)
		conn, err = kafka.DialContext(context.Background(),"tcp", kafkaAddr)
		if err != nil {
			log.Println(err)
			continue
		}

		partitions, err = conn.ReadPartitions(kafkaTopic)
		if err != nil {
			log.Println(err)
		} else {
			break
		}
	}
	defer conn.Close()


	for _, p := range partitions {
		log.Printf("'%v' reader #%v started\n", kafkaTopic, p.ID)
		go Reader([]string{kafkaAddr}, kafkaTopic, groupId, p.ID)
	}

	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir(storageDir))))

	if err := http.ListenAndServe(fmt.Sprintf(":%v", port), nil); err != nil {
		panic(err)
	}
}