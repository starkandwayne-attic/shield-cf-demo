package rand

import (
	rnd "math/rand"
	"time"
)

const charset = "0123456789abcdef"
var seed *rnd.Rand = rnd.New(rnd.NewSource(time.Now().UnixNano()))

func Bound(n, ep int) int {
	return seed.Intn(ep) + n
}

func String(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[seed.Intn(len(charset))]
	}
	return string(b)
}

func VerificationKey() string {
	return String(12)
}
