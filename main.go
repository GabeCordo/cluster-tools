package main

import (
	"ETLFramework/channel"
	"ETLFramework/cluster"
	"ETLFramework/core"
	"ETLFramework/net"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"fmt"
	"log"
	"os"
)

type Vector struct {
	x int
	y int
}

type Multiply struct {
}

func (m Multiply) ExtractFunc(output channel.OutputChannel) {
	v := Vector{1, 5} // simulate pulling data from a source
	output <- v       // send data to the TransformFunc
}

func (m Multiply) TransformFunc(input channel.InputChannel, output channel.OutputChannel) {
	if v, ok := (<-input).(Vector); ok {
		v.x = v.x * 5
		v.y = v.y * 5

		output <- v // send data to the LoadFunc
	}
}

func (m Multiply) LoadFunc(input channel.InputChannel) {
	if v, ok := (<-input).(Vector); ok {
		log.Printf("Vector(%d, %d)", v.x, v.y)
	}
}

func main() {
	// cli flags
	var debug = false

	// start reading cli arguments
	args := os.Args[1:] // strip out the file descriptor in position 0

	for i := range args {
		if args[i] == "-h" || args[i] == "--help" {
			HelpCommand()
			return
		} else if args[i] == "-d" || args[i] == "--debug" {
			debug = true
		} else if args[i] == "-g" || args[i] == "--generate-key" {
			GenerateKeyPair()
			return
		}
	}
	// stop reading cli arguments

	if debug {
		fmt.Println("starting up etlframework..")
	}

	core := core.NewCore()

	m := Multiply{}
	core.Cluster("multiply", m, cluster.Config{Identifier: "multiply"})

	core.Run()

	if debug {
		fmt.Println("shutting down etlframework")
	}
}

func HelpCommand() {
	fmt.Println("ETLFramework")
	fmt.Println("-h\tView helpful information about the etl service")
	fmt.Println("-d\tEnable debug mode")
}

func GenerateKeyPair() {
	// generate a public / private key pair
	pair, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		fmt.Println("Could not generate public and private key pair")
		return
	}

	x509Encoded, _ := x509.MarshalECPrivateKey(pair)
	fmt.Println("[private]")
	fmt.Println(net.ByteToString(x509Encoded))

	x509EncodedPub, err := x509.MarshalPKIXPublicKey(&pair.PublicKey)
	fmt.Println(len(x509EncodedPub))
	fmt.Println("[public]")
	fmt.Println(net.ByteToString(x509EncodedPub))
}
