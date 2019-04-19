package structs

import (
	"testing"
	"fmt"
	"github.com/stretchr/testify/assert"
)

func TestStructSizes(t *testing.T) {
	a := assert.New(t)

	p := &VisitPayload{
		10,
		20,
		30,
		40,
		50,
		60,
		70,
		80,
	}
	fmt.Println((len(p.Bytes())))
	fmt.Println(len(p.JSON()))

	p1 := VisitPayloadFromBytes(p.Bytes())

	a.Equal(p.DataVer, p1.DataVer)
	a.Equal(p.UserId, p1.UserId)
	a.Equal(p.EnterTime, p1.EnterTime)
	a.Equal(p.ExitTime, p1.ExitTime)
	a.Equal(p.AlgorithmType, p1.AlgorithmType)
	a.Equal(p.PoiId, p1.PoiId)
	a.Equal(p.Latitude, p1.Latitude)
	a.Equal(p.Longitude, p1.Longitude)
}