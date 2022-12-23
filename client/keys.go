package client

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/GabeCordo/commandline"
	"github.com/GabeCordo/fack"
)

// KEYS FILE

type Key struct {
	PublicKey  string
	PrivateKey string
}

type KeysFile struct {
	Keys map[string]Key
}

func (keyFile *KeysFile) ToJson(path commandline.Path) error {
	if path.DoesNotExist() {
		panic("the path is not valid, it cannot be converted to JSON")
	}

	bytes, err := json.MarshalIndent(keyFile, commandline.DefaultJSONPrefix, commandline.DefaultJSONIndent)
	if err != nil {
		panic("there was an issue marshalling the Config to JSON")
	}

	return path.Write(bytes)
}

func (keyFile *KeysFile) AddKeyPair(identity, publicKey, privateKey string) bool {
	if _, found := keyFile.Keys[identity]; found {
		// we cannot create a key with a duplicate identity
		return false
	}

	keyFile.Keys[identity] = Key{
		PublicKey:  publicKey,
		PrivateKey: privateKey,
	}

	return true
}

func (keysFile *KeysFile) RemoveKeyPair(identity string) bool {
	if _, found := keysFile.Keys[identity]; !found {
		// key identity doesn't exist, can't delete anything
		return false
	}

	delete(keysFile.Keys, identity)
	return true
}

func NewKeysFile() *KeysFile {
	keysFile := new(KeysFile)
	keysFile.Keys = make(map[string]Key)

	return keysFile
}

func JSONToKeysFile(path commandline.Path) *KeysFile {
	if path.DoesNotExist() {
		panic("keys file " + path.ToString() + " does not exist")
	}

	bytes, err := path.Read()
	if err != nil {
		panic("there was an error while reading the JSON file " + path.ToString())
	}

	keysFile := NewKeysFile()

	err = json.Unmarshal(bytes, keysFile)
	if err != nil {
		panic("there was an issue unmarshalling JSON into client.Config")
	}

	return keysFile
}

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

	keysFilePath := EtlKeysFile()
	if keysFilePath.DoesNotExist() {
		fmt.Println("etl installation is missing a keys file")
		return true
	}

	keysFile := JSONToKeysFile(keysFilePath)
	if success := keysFile.AddKeyPair(keyIdentity, x509EncodedPubStr, x509EncodedStr); !success {
		fmt.Println("failed to store key locally")
	}
	keysFile.ToJson(keysFilePath)

	return true // this is a terminal command
}

func (gkpc KeyPairCommand) DeleteKeyPair(cl *commandline.CommandLine) commandline.TerminateOnCompletion {

	keysFilePath := EtlKeysFile()
	if keysFilePath.DoesNotExist() {
		fmt.Println("missing key metadata folder, try restarting")
	}

	keyIdentity := cl.NextArg()
	if keyIdentity == commandline.FinalArg {
		fmt.Println("missing key identifier")
		return true
	}

	keysFile := JSONToKeysFile(keysFilePath)
	if success := keysFile.RemoveKeyPair(keyIdentity); !success {
		fmt.Println("key identifier does not exist")
		return true
	}

	return true // complete
}

func (gkpc KeyPairCommand) ShowKeyPairs(cl *commandline.CommandLine) commandline.TerminateOnCompletion {
	keysFilePath := EtlKeysFile()
	if keysFilePath.DoesNotExist() {
		fmt.Println("missing key metadata folder, try restarting")
	}

	keysFile := JSONToKeysFile(keysFilePath)
	for identifier, key := range keysFile.Keys {
		fmt.Printf("\nKey: %s\nPublic: %s\nPrivate: %s\n", identifier, key.PublicKey, key.PrivateKey)
	}

	return true // complete
}