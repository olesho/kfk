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
)

func Reader(kafkaBrokerUrls []string, clientId string, topic string, partition int) {
	var ctx context.Context
	var f *os.File
	var encoder *json.Encoder
	var cnt int

	var refresh = func() {
		if f != nil {
			f.Close()
		}
		var err error
		ctx, _ = context.WithTimeout(context.Background(), time.Minute*2)
		f, err = os.Create(fmt.Sprintf("%v_%v.json", partition, time.Now().Format("2006-01-02_15-04-05.999")))
		if err != nil {
			panic(err)
		}
		encoder = json.NewEncoder(f)
		cnt = 0
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

	reader := kafka.NewReader(config)
	defer reader.Close()

	refresh()
	for {
		m, err := reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				fmt.Println("deadline exceeded")
				refresh()
			} else {
				fmt.Println("error while receiving message: ", err.Error())
			}
		} else {
			p := structs.VisitPayloadFromBytes(m.Value)
			err = encoder.Encode(p)
			if err != nil {
				fmt.Println("error while encoding message: ", err.Error())
			}
			cnt++
			fmt.Println(cnt)
			if cnt == 100 {
				refresh()
				atomic.AddUint32(globalCnt, 100)
			}
		}
	}
	f.Close()
}

var kafkaAddr = os.Getenv("KAFKA_ADDR")

var globalCnt *uint32 = new(uint32)

func main() {
	conn, err := kafka.Dial("tcp", kafkaAddr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	partitions, err := conn.ReadPartitions("visit")
	if err != nil {
		panic(err)
	}
	for _, p := range partitions {
		fmt.Println("Reader", p.ID, "started	")
		go Reader([]string{kafkaAddr}, "1", "visit", p.ID)
	}

	t := time.NewTicker(time.Second)
	for {
		<- t.C

		fmt.Println(*globalCnt)
	}
	//wg := sync.WaitGroup{}
	//wg.Add(1)
	//wg.Wait()
}