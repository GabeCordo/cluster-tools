package client

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"fmt"
	"github.com/GabeCordo/commandline"
	"github.com/GabeCordo/fack"
)

// GENERATE KEY PAIR START

type GenerateKeyPairCommand struct {
	PublicName string
}

func (gkpc GenerateKeyPairCommand) Name() string {
	return gkpc.PublicName
}

func (gkpc GenerateKeyPairCommand) Run(cl *commandline.CommandLine) commandline.TerminateOnCompletion {

	// we only want to create a key if it has been pre-pended with the 'create' flag
	if cl.Flags.Create {
		// generate a public / private key pair
		pair, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			fmt.Println("Could not generate public and private key pair")
			return true
		}

		x509Encoded, _ := x509.MarshalECPrivateKey(pair)
		fmt.Println("[private]")
		fmt.Println(fack.ByteToString(x509Encoded))

		x509EncodedPub, err := x509.MarshalPKIXPublicKey(&pair.PublicKey)
		fmt.Println(len(x509EncodedPub))
		fmt.Println("[public]")
		fmt.Println(fack.ByteToString(x509EncodedPub))
	} else {
		fmt.Println("key specified without an action [create/delete]?")
	}

	return true // this is a terminal command
}
