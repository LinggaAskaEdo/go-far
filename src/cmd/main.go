package main

import (
	"time"

	"go-far/src/util"
)

const (
	DefaultMinJitter = 100
	DefaultMaxJitter = 2000
)

func sleepWithJitter(low, high int) {
	if low < 1 {
		low = DefaultMinJitter
	}

	if high < 1 || high < low {
		high = DefaultMaxJitter
	}

	rnd := util.RandomInt(high-low) + low
	time.Sleep(time.Duration(rnd) * time.Millisecond)
}
