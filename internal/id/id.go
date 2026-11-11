package id

import (
	"crypto/rand"
	"encoding/hex"
)

// New returns a short random id with the given prefix.
func New(prefix string) string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err)
	}
	return prefix + hex.EncodeToString(b[:])
}
