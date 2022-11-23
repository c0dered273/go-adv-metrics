package metrics

import (
	"math/rand"
	"time"
)

var pollCounter int64

func GetCustomCounter() []Counter {
	pollCounter += 1
	return []Counter{
		{
			name:  "PollCounter",
			value: pollCounter,
		},
	}
}

func GetCustomGauge() []Gauge {
	seed := rand.NewSource(time.Now().UnixNano())
	r := rand.New(seed)
	return []Gauge{
		{
			name:  "RandomValue",
			value: r.Float64() * 1000000,
		},
	}
}
