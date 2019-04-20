package main

import (
	"github.com/segmentio/kafka-go"
	"time"
	"sync"
	"fmt"
	"context"
)

type Gateway struct {
	messages chan kafka.Message
	writer *kafka.Writer
	timer *time.Timer
	slice *[]kafka.Message
	lock *sync.Mutex
}

func NewGateway(addr, topic, clientId string) *Gateway {
	dialer := &kafka.Dialer{
		Timeout:  10 * time.Second,
		ClientID: clientId,
	}

	return &Gateway{
		messages : make(chan kafka.Message, 500),
		writer : kafka.NewWriter(kafka.WriterConfig{
			BatchSize:		  1,
			Brokers:          []string{addr},
			Topic:            topic,
			Dialer:           dialer,
			WriteTimeout:     10 * time.Second,
			ReadTimeout:      10 * time.Second,
		}),
		timer : time.NewTimer(time.Second),
		slice : new([]kafka.Message),
		lock : &sync.Mutex{},
	}
}

func (v *Gateway) send() {
	v.lock.Lock()
	defer v.lock.Unlock()
	if len(*v.slice) > 0 {
		err := v.writer.WriteMessages(context.Background(), *v.slice...)
		if err != nil {
			fmt.Println(err)
		}
		v.timer.Reset(time.Second)
		*v.slice = make([]kafka.Message, 0)
	}
}

func (v *Gateway) Push(m *kafka.Message) {
	v.messages <- *m
}

func (v *Gateway) RunStream() {
	go func () {
		for {
			*v.slice = append(*v.slice, <- v.messages)
			if len(*v.slice) == 500 {
				v.send()
			}
		}
	}()

	go func() {
		for {
			<- v.timer.C
			v.send()
		}
	}()
}

func (v *Gateway) Close() error {
	v.lock.Lock()
	defer v.lock.Unlock()
	v.timer.Stop()
	return v.writer.Close()
}
