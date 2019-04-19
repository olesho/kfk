package main

import (
	"time"
	"github.com/olesho/hl/structs"
	"math/rand"
	"encoding/json"
	"sync/atomic"
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
		Longitude:		0,
	}
	vg.currentStartTime = vg.currentStartTime.Add(time.Hour*2)
	return p
}

func (vg *VisitGenerator) CurrentDay() int {
	return vg.currentStartTime.Day()
}

func (vg * VisitGenerator) Run(amount int64) {
	var payload []byte
	var num = new(int64)
	*num = amount

	for {
		if *num == 0 {
			break
		}

		oldDay := vg.CurrentDay()
		data, _ := json.Marshal(vg.Next())
		data = append(data, []byte("\n")...)
		payload = append(payload, data...)
		atomic.AddInt64(num, -1)

		if oldDay != vg.CurrentDay() {
			var buf bytes.Buffer
			g := gzip.NewWriter(&buf)
			if _, err := g.Write(data); err != nil {
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

			payload = []byte{}
		}
	}

	if len(payload) > 0 {
		req, err := http.NewRequest("POST", vg.endpoint, bytes.NewReader(payload))
		if err != nil {
			vg.logger.Println(err)
			return
		}

		_, err = http.DefaultClient.Do(req)
		if err != nil {
			vg.logger.Println(err)
		}
	}

	vg.done <- struct{}{}
}