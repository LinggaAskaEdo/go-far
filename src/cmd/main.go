package main

import (
	"math/rand"
	"time"
)

const (
	DefaultMinJitter = 100
	DefaultMaxJitter = 2000
)

func sleepWithJitter(min int, max int) {
	if min < 1 {
		min = DefaultMinJitter
	}

	if max < 1 || max < min {
		max = DefaultMaxJitter
	}

	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	rnd := rand.Intn(max-min) + min
	time.Sleep(time.Duration(rnd) * time.Millisecond)
}
