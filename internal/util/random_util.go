package util

import (
	"crypto/rand"
	"math/big"
)

// RandomInt returns a cryptographically secure random int in [0, n).
func RandomInt(n int) int {
	if n <= 0 {
		return 0
	}

	m, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		panic(err) // crypto/rand should never fail
	}

	return int(m.Int64())
}

// randomFloat32 returns a cryptographically secure float32 in [0.0, 1.0).
func RandomFloat32() float32 {
	// Generate a random uint32 and convert to float32.
	// This gives 2^32 equally spaced values in [0,1).
	var buf [4]byte

	_, err := rand.Read(buf[:])
	if err != nil {
		panic(err)
	}

	// Interpret as big-endian uint32
	u := uint32(buf[0])<<24 | uint32(buf[1])<<16 | uint32(buf[2])<<8 | uint32(buf[3])
	return float32(u) / (1 << 32)
}

// shuffleSlice uses crypto/rand to shuffle any slice in place.
func ShuffleSlice[T any](slice []T) {
	n := len(slice)
	for i := n - 1; i > 0; i-- {
		j := RandomInt(i + 1)
		slice[i], slice[j] = slice[j], slice[i]
	}
}
