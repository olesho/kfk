package structs

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
)

const VISIT_SIZE = 56

type Payload interface {
	Bytes() []byte
}

type VisitPayload struct {
	DataVer int
	UserId  int64
	EnterTime  int64
	ExitTime  int64
	AlgorithmType  int
	PoiId  int64
	Latitude  int64
	Longitude  int64
}

func VisitPayloadFromBytes(b []byte) (*VisitPayload, error) {
	if len(b) == 56 {
		p := &VisitPayload{}
		p.DataVer = int(binary.BigEndian.Uint32(b[:4]))
		p.UserId = int64(binary.BigEndian.Uint64(b[4:12]))
		p.EnterTime = int64(binary.BigEndian.Uint64(b[12:20]))
		p.ExitTime = int64(binary.BigEndian.Uint64(b[20:28]))
		p.AlgorithmType = int(binary.BigEndian.Uint32(b[28:32]))
		p.PoiId = int64(binary.BigEndian.Uint64(b[32:40]))
		p.Latitude = int64(binary.BigEndian.Uint64(b[40:48]))
		p.Longitude = int64(binary.BigEndian.Uint64(b[48:56]))
		return p, nil
	} else {
		return nil, errors.New(fmt.Sprintf("Error decoding 'visit': invalid payload size: %v", len(b)))
	}
}

func (p *VisitPayload) Bytes() []byte {
	b := make([]byte, 56)
	binary.BigEndian.PutUint32(b[:4], uint32(p.DataVer))
	binary.BigEndian.PutUint64(b[4:12], uint64(p.UserId))
	binary.BigEndian.PutUint64(b[12:20], uint64(p.EnterTime))
	binary.BigEndian.PutUint64(b[20:28], uint64(p.ExitTime))
	binary.BigEndian.PutUint32(b[28:32], uint32(p.AlgorithmType))
	binary.BigEndian.PutUint64(b[32:40], uint64(p.PoiId))
	binary.BigEndian.PutUint64(b[40:48], uint64(p.Latitude))
	binary.BigEndian.PutUint64(b[48:56], uint64(p.Longitude))
	return b
}

func (p *VisitPayload) JSON() []byte {
	r, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}
	return r
}

type ActivityPayload struct {
	DataVer 		int
	UserId  		int64
	ActivityType 	int
	StartTime  		int64
	EndTime  		int64
	StartLatitude	int64
	StartLongitude 	int64
	EndLatitude		int64
	EndLongitude 	int64
}

func ActivityPayloadFromBytes(b []byte) (*ActivityPayload, error) {
	if len(b) == 64 {
		p := &ActivityPayload{}
		p.DataVer = int(binary.BigEndian.Uint32(b[:4]))
		p.UserId = int64(binary.BigEndian.Uint64(b[4:12]))
		p.ActivityType = int(binary.BigEndian.Uint32(b[12:16]))
		p.StartTime = int64(binary.BigEndian.Uint64(b[16:24]))
		p.EndTime = int64(binary.BigEndian.Uint32(b[24:32]))
		p.StartLatitude = int64(binary.BigEndian.Uint64(b[32:40]))
		p.StartLongitude = int64(binary.BigEndian.Uint64(b[40:48]))
		p.EndLatitude = int64(binary.BigEndian.Uint64(b[48:56]))
		p.EndLongitude = int64(binary.BigEndian.Uint64(b[56:64]))
		return p, nil
	} else {
		return nil, errors.New(fmt.Sprintf("Error decoding 'activity': invalid payload size: %v", len(b)))
	}
}

func (p *ActivityPayload) Bytes() []byte {
	b := make([]byte, 64)
	binary.BigEndian.PutUint32(b[:4], uint32(p.DataVer))
	binary.BigEndian.PutUint64(b[4:12], uint64(p.UserId))
	binary.BigEndian.PutUint32(b[12:16], uint32(p.ActivityType))
	binary.BigEndian.PutUint64(b[16:24], uint64(p.StartTime))
	binary.BigEndian.PutUint64(b[24:32], uint64(p.EndTime))
	binary.BigEndian.PutUint64(b[32:40], uint64(p.StartLatitude))
	binary.BigEndian.PutUint64(b[40:48], uint64(p.StartLongitude))
	binary.BigEndian.PutUint64(b[48:56], uint64(p.EndLatitude))
	binary.BigEndian.PutUint64(b[56:64], uint64(p.EndLongitude))
	return b
}

func (p *ActivityPayload) JSON() []byte {
	r, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}
	return r
}

func RandLatitude() int64 {
	return rand.Int63n(36000)-18000
}

func RandLongitude() int64 {
	return rand.Int63n(9000)
}

func Int64ToGUUD(l int64) float64 {
	return float64(l)/100
}
