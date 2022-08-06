package net

import (
	"bytes"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

// integer constants
const (
	maxGeneratedStringLength = 100
	lowerASCIIBound          = 97
	upperASCIIBound          = 122
	x509ByteLength           = 91
)

// string constants
const (
	EmptyString = ""
	StringSpace = " "
)

func RandInteger(min int, max int) int {
	rand.Seed(time.Now().UTC().UnixNano())
	return min + rand.Intn(max-min)
}

func RandInteger64(min int64, max int64) int64 {
	rand.Seed(time.Now().UTC().UnixNano())
	return min + rand.Int63n(max-min)
}

func GenerateRandomString(seed int) string {
	buffer := new(bytes.Buffer)
	for i := 0; i < maxGeneratedStringLength; i++ {
		char := RandInteger(lowerASCIIBound, upperASCIIBound)
		buffer.WriteString(string(char))
	}
	return buffer.String()
}

func GenerateNonce() int64 {
	return time.Now().Unix() * RandInteger64(4, 9)
}

func GetInternetProtocol(r *http.Request) (*Address, error) {
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return NewAddress(forwarded)
	}
	return NewAddress(r.RemoteAddr)
}

func IsUsingJSONContent(r *http.Request) bool {
	content := r.Header.Get("Content-Type")
	return content == "application/json"
}

func ArrayToLookupHashTable(a []string) map[string]any {
	m := make(map[string]any)
	for _, i := range a {
		m[i] = nil
	}
	return m
}

func ByteToString(data []byte) string {
	var s string

	for i := range data {
		s += strconv.FormatInt(int64(data[i]), 10)
		if i != (len(data) - 1) {
			s += " "
		}
	}

	return s
}

// TODO - clean up this function
func StringToByte(data string) ([]byte, bool) {
	b := [x509ByteLength]byte{}

	var s string

	j := 0
	chars := []rune(data)
	for i := 0; i < len(data); i++ {
		char := string(chars[i])
		if char != StringSpace {
			s += char
		} else {
			val, err := strconv.Atoi(s)
			if err != nil {
				return b[:], false
			}
			b[j] = byte(val)
			s = EmptyString
			j++
		}
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return b[:], false
	}
	b[j] = byte(val)

	return b[:], true
}
