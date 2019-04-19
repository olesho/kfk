package main

import (
	"github.com/segmentio/kafka-go"
	"time"
	"fmt"
	"context"
)

func DeleteTopic(topic string) error {
	conn, err := kafka.Dial("tcp", kafkaAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	return conn.DeleteTopics(topic)
}

func CreateTopic(topic string, numPartitions int) error {
	conn, err := kafka.Dial("tcp", kafkaAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	return conn.CreateTopics(kafka.TopicConfig{
		Topic: topic,
		NumPartitions:	numPartitions,
		ReplicationFactor: 1,
	})
}

func ListPartitions(topic string) ([]int, error) {
	conn, err := kafka.Dial("tcp", kafkaAddr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	partitions, err := conn.ReadPartitions(topic)
	if err != nil {
		return nil, err
	}

	res := []int{}
	for _, p := range partitions {
		res = append(res, p.ID)
	}
	return res, nil
}

func ConfigureBatchWriter(kafkaBrokerUrls []string, clientId string, topic string) (w *kafka.Writer, err error) {
	dialer := &kafka.Dialer{
		Timeout:  10 * time.Second,
		ClientID: clientId,
	}

	config := kafka.WriterConfig{
		BatchSize:		  100,
		Brokers:          kafkaBrokerUrls,
		Topic:            topic,
		//Balancer:		&kafka.Hash{},
		//Balancer:         &kafka.LeastBytes{},
		Dialer:           dialer,
		WriteTimeout:     10 * time.Second,
		ReadTimeout:      10 * time.Second,
		//CompressionCodec: snappy.NewCompressionCodec(),
	}
	w = kafka.NewWriter(config)
	return w, nil
}

func Reader(kafkaBrokerUrls []string, clientId string, topic string) {
	config := kafka.ReaderConfig{
		Brokers:         kafkaBrokerUrls,
		GroupID:         clientId,
		Topic:           topic,
		MinBytes:        10e3,            // 10KB
		MaxBytes:        10e6,            // 10MB
		MaxWait:         1 * time.Second, // Maximum amount of time to wait for new data to come when fetching batches of messages from kafka.
		ReadLagInterval: -1,
	}

	reader := kafka.NewReader(config)
	defer reader.Close()

	for {
		m, err := reader.ReadMessage(context.Background())
		if err != nil {
			fmt.Println("error while receiving message: ", err.Error())
			continue
		}

		fmt.Printf("message at topic/partition/offset %v/%v/%v: %s\n", m.Topic, m.Partition, m.Offset, string(m.Value))
	}
}