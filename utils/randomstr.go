package utils

import (
	"math/rand"
	"time"
)

// charset of all characters that can be used in the random string
const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!$%&#"

/*
RandString returns a randomly generated string
of the given length which can contain
lower and upper case letters, digits aswell as
the special characters !, $, %, &, #
*/
func RandString(length int) string {
	// Set seed to current timestamp
	rand.Seed(time.Now().UnixNano())

	// Create random []byte with letters from the charset
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}

	return string(b)
}
