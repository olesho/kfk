package main

import (
	"time"
	"github.com/olesho/kfk/structs"
	"math/rand"
	"encoding/json"
	"net/http"
	"bytes"
	"log"
	"compress/gzip"
)

type VisitGenerator struct {
	clientId 			int64
	currentStartTime	time.Time
	endpoint 			string
	done				chan struct{}
	logger 				*log.Logger

	_cnt int64
}

func NewVisitGenerator(clientId int64, endpoint string, doneSignal chan struct{}, logger *log.Logger) *VisitGenerator {
	startTime, _ := time.Parse("2006-01-02T15:04:05", "2018-01-01T00:00:00")
	return &VisitGenerator{
		clientId: 			clientId,
		currentStartTime: 	startTime,
		endpoint:			endpoint,
		done:				doneSignal,
		logger:				logger,
	}
}

func (vg *VisitGenerator) Next() *structs.VisitPayload {
	p := &structs.VisitPayload{
		DataVer:		1,
		UserId:			vg.clientId,
		EnterTime:		vg.currentStartTime.UnixNano(),
		ExitTime:		vg.currentStartTime.Add(time.Minute*40).UnixNano(),
		AlgorithmType: 	rand.Intn(6)+1,
		PoiId:         	rand.Int63(),
		Latitude:		0,
		Longitude:		vg._cnt,
	}
	vg._cnt++
	vg.currentStartTime = vg.currentStartTime.Add(time.Hour*2)
	return p
}

func (vg *VisitGenerator) CurrentDay() int {
	return vg.currentStartTime.Day()
}

func (vg * VisitGenerator) Run(amount int) {
	var payload []byte

	for i := 0; i < amount; i++ {
		oldDay := vg.CurrentDay()
		data, _ := json.Marshal(vg.Next())
		data = append(data, []byte("\n")...)
		payload = append(payload, data...)

		if oldDay != vg.CurrentDay() {
			vg.send(payload)
			payload = []byte{}
		}
	}
	vg.send(payload)

	vg.done <- struct{}{}
}

func (vg * VisitGenerator) send(payload []byte) {
	if len(payload) > 0 {
		//req, err := http.NewRequest("POST", vg.endpoint, bytes.NewBuffer(payload))

		var buf bytes.Buffer
		g := gzip.NewWriter(&buf)
		if _, err := g.Write(payload); err != nil {
			vg.logger.Println(err)
			return
		}
		if err := g.Close(); err != nil {
			vg.logger.Println(err)
			return
		}

		req, err := http.NewRequest("POST", vg.endpoint, &buf)
		if err != nil {
			vg.logger.Println(err)
			return
		}
		req.Header.Add("Content-Encoding", "gzip")
		_, err = http.DefaultClient.Do(req)
		if err != nil {
			vg.logger.Println(err)
		}
	}
}