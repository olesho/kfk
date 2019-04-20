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
)

func Reader(kafkaBrokerUrls []string, topic string, partition int) {
	var ctx context.Context
	var f *os.File
	var encoder *json.Encoder
	var cnt = new(uint32)
	var fileName string
	var reader *kafka.Reader

	var refresh = func() {
		if f != nil {
			err := f.Close()
			if err != nil {
				log.Println(err)
			}
			log.Println(fmt.Sprintf("%v records written to %v", *cnt, fileName))
			log.Printf("Current offset: %v\n", reader.Offset())
		}
		var err error
		ctx, _ = context.WithTimeout(context.Background(), time.Minute*2)
		fileName = fmt.Sprintf("%v_%v_%v.json", topic, partition, time.Now().Format("2006-01-02_15-04-05.999"))
		f, err = os.Create(fileName)
		if err != nil {
			log.Println(err)
			return
		}
		encoder = json.NewEncoder(f)
		*cnt = 0
	}

	config := kafka.ReaderConfig{
		Partition:		partition,
		Brokers:		kafkaBrokerUrls,
		//GroupID:         clientId,
		Topic:			topic,
		MinBytes:		10e3,            // 10KB
		MaxBytes:		10e6,            // 10MB
		MaxWait:		1 * time.Second, // Maximum amount of time to wait for new data to come when fetching batches of messages from kafka.
		ReadLagInterval:	-1,
	}

	reader = kafka.NewReader(config)
	defer reader.Close()

	refresh()
	for {
		m, err := reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				log.Println("deadline exceeded")
				refresh()
			} else {
				log.Println("error while receiving message: ", err.Error())
			}
		} else {
			if kafkaTopic == "visit" {
				p, err := structs.VisitPayloadFromBytes(m.Value)
				if err != nil {
					log.Println(err)
				} else {
					err = encoder.Encode(p)
					if err != nil {
						log.Println("error while encoding message: ", err.Error())
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
					}
				}
			}

			atomic.AddUint32(cnt, 1)
			if *cnt == 100 {
				refresh()
				atomic.AddUint32(globalCnt, 100)
			}
		}
	}
	err := f.Close()
	if err != nil {
		log.Println(err)
	}
	log.Println(fmt.Sprintf("%v records written to %v", *cnt, fileName))
}

var kafkaAddr = os.Getenv("KAFKA_ADDR")
var kafkaTopic = os.Getenv("TOPIC")

var globalCnt = new(uint32)

func main() {
	// delay
	time.Sleep(time.Second*2)

	conn, err := kafka.DialContext(context.Background(),"tcp", kafkaAddr)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	partitions, err := conn.ReadPartitions(kafkaTopic)
	if err != nil {
		log.Println(err)
		return
	}
	for _, p := range partitions {
		log.Printf("'%v' reader #%v started\n", kafkaTopic, p.ID)
		go Reader([]string{kafkaAddr}, kafkaTopic, p.ID)
	}


	t := time.NewTicker(time.Second)
	for {
		<- t.C
	}
}