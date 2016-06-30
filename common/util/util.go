package util

import (
	"math/rand"
	"time"
)

func RandomString(num int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	var temp []byte = make([]byte, num)
	for i := 0; i < num; i++ {
		temp[i] = byte(33 + rand.Intn(93))
	}
	return string(temp[:])
}

func RandomInt(min int, max int) int {
	rand.Seed(time.Now().UTC().UnixNano())
	return min + rand.Intn(max-min)
}
