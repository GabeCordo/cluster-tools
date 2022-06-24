package main

import (
	"ETLFramework/net"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"net/http"
	"testing"
	"time"
)

/*? Test Function */

func AuthenticatedIndex(response *net.Response) {
	response.AddStatus(http.StatusOK, "authenticated")
}

func TestAuthNoGlobalOrLocalPermissionsPresent(t *testing.T) {

}

func TestAuthGlobalPermissionPresent(t *testing.T) {

}

func TestAuthLocalPermissionPresent(t *testing.T) {

}

func TestAuthGlobalAndLocalPermissionsPresent(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Error("Could not generate an ECDSA key pair")
	}

	var na *net.Auth = net.NewAuth()
	var ne *net.Endpoint = net.NewEndpoint("test", &privateKey.PublicKey)

	globalPermissionMap := net.NewPermission(false, false, false, false)
	localPermissionMap := net.NewPermission(true, false, false, false)
	ne.AddGlobalPermission(globalPermissionMap)
	ne.AddLocalPermission("/", localPermissionMap)
	na.AddTrusted("127.0.0.1", ne)

	a := net.Address{"localhost", 8000}
	n := net.NewNode("", a, false, na, nil) // pass a nil to logger pointer ~ no logging
	n.Route("/", AuthenticatedIndex, []string{"GET", "POST"}, true)

	go n.Start()

	// if you are on macos, you may need to give the binary permission to use a socket port
	time.Sleep(WaitForServerStart)

	request := net.NewRequest("/")
	request.Sign(privateKey)

	/* GET request should succeed */
	resp, err := request.Send("GET", LocalHost+fmt.Sprint(GETPort))
	if err != nil {
		t.Error("Failed to startup an HTTP GET route.")
	}

	if (*resp).Data["status"] != "authenticated" {
		t.Error("Node could not authenticate a valid host")
	}

	/* POST request should NOT succeed */
	resp, err = request.Send("POST", LocalHost+fmt.Sprint(GETPort))
	if err != nil {
		t.Error("Failed to startup an HTTP GET route.")
	}

	if resp.Status == http.StatusUnauthorized {
		t.Error("Node was let into a permission ")
	}

}
