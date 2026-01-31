package utils

import (
	"fmt"
	"math/rand"
	"time"
)

func ShortID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}

func GenerateWorkerName(prefix string) string {
	chars := "abcdefghijklmnopqrstuvwxyz0123456789"
	suffix := make([]byte, 6)
	for i := range suffix {
		suffix[i] = chars[rand.Intn(len(chars))]
	}
	return fmt.Sprintf("%s%d-%s", prefix, time.Now().Unix(), string(suffix))
}
