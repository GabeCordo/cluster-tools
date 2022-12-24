package core

import (
	"encoding/json"
	"github.com/GabeCordo/commandline"
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
