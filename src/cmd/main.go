package main

import (
	"math/rand"
	"sync"
	"time"
)

const (
	DefaultMinJitter = 100
	DefaultMaxJitter = 2000
)

var (
	defaultRand     = rand.New(rand.NewSource(time.Now().UnixNano()))
	defaultRandOnce sync.Once
)

func sleepWithJitter(min int, max int) {
	if min < 1 {
		min = DefaultMinJitter
	}

	if max < 1 || max < min {
		max = DefaultMaxJitter
	}

	defaultRandOnce.Do(func() {
		defaultRand = rand.New(rand.NewSource(time.Now().UnixNano()))
	})

	rnd := defaultRand.Intn(max-min) + min
	time.Sleep(time.Duration(rnd) * time.Millisecond)
}
