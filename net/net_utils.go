package net

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
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

func GetInternetProtocol(r *http.Request) string {
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
}

func FormatResponse(w http.ResponseWriter, httpResponseCode int, data string) string {
	response := fmt.Sprintf("{\"status\":%d, \"data\": %s }", httpResponseCode, data)
	fmt.Println(response)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpResponseCode)
	json.NewEncoder(w).Encode(response)
	return response
}

func IsUsingJSONContent(r *http.Request) bool {
	content := r.Header.Get("Content-Type")
	return content == "application/json"
}
