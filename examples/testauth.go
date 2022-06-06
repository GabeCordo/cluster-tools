package main

import (
	"ETLFramework/core"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	mrand "math/rand"
	"strconv"
	"time"
)

func main() {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}
	mrand.Seed(time.Now().UnixNano())
	nonce := 1 + mrand.Intn(9)
	msg := "/extract" + strconv.Itoa(nonce)
	hash := sha256.Sum256([]byte(msg))

	sig, err := ecdsa.SignASN1(rand.Reader, privateKey, hash[:])
	if err != nil {
		panic(err)
	}

	var na core.NodeAuth = NewAuth()
	var ne core.NodeEndpoint = core.NewEndpoint("test", [4]bool{true, true, true, true}, privateKey.PublicKey)
	na.AddTrusted("10.10.10.1", &ne)

	fmt.Println("core auth verified: ", na.ValidateSource("10.10.10.1", hash[:], sig))

	valid := ecdsa.VerifyASN1(&privateKey.PublicKey, hash[:], sig)
	fmt.Println("signature verified: ", valid)

}
