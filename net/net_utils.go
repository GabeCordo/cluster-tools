package net

import (
	"bytes"
	"fmt"
	"math/rand"
	"time"
)

const (
	maxGeneratedStringLength = 100
	lowerASCIIBound          = 97
	upperASCIIBound          = 122
)

func RandInteger(min int, max int) int {
	rand.Seed(time.Now().UTC().UnixNano())
	return min + rand.Intn(max-min)
}

func GenerateRandomString(seed int) string {
	buffer := new(bytes.Buffer)
	for i := 0; i < maxGeneratedStringLength; i++ {
		char := RandInteger(lowerASCIIBound, upperASCIIBound)
		buffer.WriteString(string(char))
	}
	return buffer.String()
}

func main() {
	s := GenerateRandomString(24)
	fmt.Println(s)
}
