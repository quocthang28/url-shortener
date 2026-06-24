package shortener

import (
	"crypto/rand"
)

// base62 set: a-z, A-Z, 0-9.
const Alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// Generate returns a cryptographically random base62 string of length n.
func Generate(n int) (string) {
	buf := make([]byte, n)

	_, _ = rand.Read(buf)

	for i, b := range buf {
		buf[i] = Alphabet[int(b)%len(Alphabet)]
	}

	return string(buf)
}
