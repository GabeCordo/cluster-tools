package cli

import (
	"ETLFramework/net"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func HelpCommand() {
	fmt.Println("ETLFramework")
	fmt.Println("-h\tView helpful information about the etl service")
	fmt.Println("-d\tEnable debug mode")
	fmt.Println("-g\tGenerate an ECDSA x509 public and private key pair")
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

func InteractiveDashboard() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs // block until we receive an interrupt from the system
		fmt.Println()
		os.Exit(0)
	}()

	for {
		fmt.Printf("time: %s\r", time.Now())
	}
}
