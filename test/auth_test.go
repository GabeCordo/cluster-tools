package main

import (
	"ETLFramework/net"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	mrand "math/rand"
	"strconv"
	"testing"
	"time"
)

/*? Test Function */

func TestNodeAuth(t *testing.T) {
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
	fmt.Println("----------Signature----------")
	fmt.Printf("%v\n", sig)
	fmt.Printf("%d", len(sig))

	var na net.NodeAuth = net.NewAuth()
	var pm net.Permission = net.NewPermission(true, true, true, true)
	var ne net.NodeEndpoint = net.NewEndpoint("test", &pm, &privateKey.PublicKey)
	na.AddTrusted("10.10.10.1", &ne)

	fmt.Println("net auth verified: ", na.ValidateSource("10.10.10.1", hash[:], sig))

	valid := ecdsa.VerifyASN1(&privateKey.PublicKey, hash[:], sig)

	if !valid {
		t.Error("ECDSA signature/public-key verification failed!")
	}
}
