package main

import (
	"time"
	"github.com/olesho/kfk/structs"
	"math"
	"math/rand"
	"encoding/json"
	"net/http"
	"bytes"
	"log"
	"compress/gzip"
)

type ActivityGenerator struct {
	clientId 			int64
	currentStartTime	time.Time
	endpoint 			string
	done				chan struct{}
	logger 				*log.Logger
}

func NewActivityGenerator(clientId int64, endpoint string, doneSignal chan struct{}, logger *log.Logger) *ActivityGenerator {
	startTime, _ := time.Parse("2006-01-02T15:04:05", "2018-01-01T00:00:00")
	return &ActivityGenerator{
		clientId: 			clientId,
		currentStartTime: 	startTime,
		endpoint:			endpoint,
		done:				doneSignal,
		logger:				logger,
	}
}

func (ag *ActivityGenerator) Next() *structs.ActivityPayload {
	p := &structs.ActivityPayload{
		DataVer: 		1,
		UserId: 		ag.clientId,
		ActivityType: 	int(math.Pow(2, float64(rand.Intn(31)+1))),
		StartTime:		ag.currentStartTime.UnixNano(),
		EndTime:		ag.currentStartTime.Add(time.Minute*30).UnixNano(),
		StartLatitude:	0,
		StartLongitude:	0,
		EndLatitude:	0,
		EndLongitude:	0,
	}
	ag.currentStartTime = ag.currentStartTime.Add(time.Hour*2)
	return p
}

func (ag *ActivityGenerator) CurrentDay() int {
	return ag.currentStartTime.Day()
}

func (ag *ActivityGenerator) Run(amount int) {
	var payload []byte

	for i := 0; i < amount; i++ {
		oldDay := ag.CurrentDay()
		data, _ := json.Marshal(ag.Next())
		data = append(data, []byte("\n")...)
		payload = append(payload, data...)

		if oldDay != ag.CurrentDay() {
			ag.send(payload)
			payload = []byte{}
		}
	}
	ag.send(payload)

	ag.done <- struct{}{}
}

func (ag * ActivityGenerator) send(payload []byte) {
	if len(payload) > 0 {
		//req, err := http.NewRequest("POST", ag.endpoint, bytes.NewBuffer(payload))

		var buf bytes.Buffer
		g := gzip.NewWriter(&buf)
		if _, err := g.Write(payload); err != nil {
			ag.logger.Println(err)
			return
		}
		if err := g.Close(); err != nil {
			ag.logger.Println(err)
			return
		}

		req, err := http.NewRequest("POST", ag.endpoint, &buf)
		if err != nil {
			ag.logger.Println(err)
			return
		}
		req.Header.Add("Content-Encoding", "gzip")
		_, err = http.DefaultClient.Do(req)
		if err != nil {
			ag.logger.Println(err)
		}
	}
}