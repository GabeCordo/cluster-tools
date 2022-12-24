package controllers

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"fmt"
	"github.com/GabeCordo/commandline"
	"github.com/GabeCordo/etl/client/core"
	"github.com/GabeCordo/fack"
)

// GENERATE KEY PAIR START

type KeyPairCommand struct {
	PublicName string
}

func (gkpc KeyPairCommand) Name() string {
	return gkpc.PublicName
}

func (gkpc KeyPairCommand) Run(cl *commandline.CommandLine) commandline.TerminateOnCompletion {
	if cl.Flags.Create {
		gkpc.GenerateKeyPair(cl)
	} else if cl.Flags.Delete {
		gkpc.DeleteKeyPair(cl)
	} else if cl.Flags.Show {
		gkpc.ShowKeyPairs(cl)
	}

	return true // complete
}

func (gkpc KeyPairCommand) GenerateKeyPair(cl *commandline.CommandLine) commandline.TerminateOnCompletion {

	keyIdentity := cl.NextArg()
	if keyIdentity == commandline.FinalArg {
		keyIdentity = fack.GenerateRandomString(10) // seed is a randomly defined value
	}

	// generate a public / private key pair
	pair, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		fmt.Println("Could not generate public and private key pair")
		return true
	}

	x509Encoded, _ := x509.MarshalECPrivateKey(pair)
	x509EncodedStr := fack.ByteToString(x509Encoded)
	fmt.Println("[private]")
	fmt.Println(x509EncodedStr)

	x509EncodedPub, err := x509.MarshalPKIXPublicKey(&pair.PublicKey)
	x509EncodedPubStr := fack.ByteToString(x509EncodedPub)
	fmt.Println("[public]")
	fmt.Println(x509EncodedPubStr)

	keysFilePath := core.EtlKeysFile()
	if keysFilePath.DoesNotExist() {
		fmt.Println("etl installation is missing a keys file")
		return true
	}

	keysFile := core.JSONToKeysFile(keysFilePath)
	if success := keysFile.AddKeyPair(keyIdentity, x509EncodedPubStr, x509EncodedStr); !success {
		fmt.Println("failed to store key locally")
	}
	keysFile.ToJson(keysFilePath)

	return true // this is a terminal command
}

func (gkpc KeyPairCommand) DeleteKeyPair(cl *commandline.CommandLine) commandline.TerminateOnCompletion {

	keysFilePath := core.EtlKeysFile()
	if keysFilePath.DoesNotExist() {
		fmt.Println("missing key metadata folder, try restarting")
	}

	keyIdentity := cl.NextArg()
	if keyIdentity == commandline.FinalArg {
		fmt.Println("missing key identifier")
		return true
	}

	keysFile := core.JSONToKeysFile(keysFilePath)
	if success := keysFile.RemoveKeyPair(keyIdentity); !success {
		fmt.Println("key identifier does not exist")
		return true
	}

	return true // complete
}

func (gkpc KeyPairCommand) ShowKeyPairs(cl *commandline.CommandLine) commandline.TerminateOnCompletion {
	keysFilePath := core.EtlKeysFile()
	if keysFilePath.DoesNotExist() {
		fmt.Println("missing key metadata folder, try restarting")
		return true
	}

	keysFile := core.JSONToKeysFile(keysFilePath)
	for identifier, key := range keysFile.Keys {
		fmt.Printf("\nKey: %s\nPublic: %s\nPrivate: %s\n", identifier, key.PublicKey, key.PrivateKey)
	}

	return true // complete
}
