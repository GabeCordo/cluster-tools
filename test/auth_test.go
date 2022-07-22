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

func AuthenticatedIndex(request *net.Request, response *net.Response) {
	response.AddStatus(http.StatusOK, "authenticated")
}

// since there is no global or local Permission bitmap assigned to the endpoint
// the auth function should not grant access to the route
func TestAuthNoGlobalOrLocalPermissionsPresent(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Error("Could not generate an ECDSA key pair")
	}

	var na *net.Auth = net.NewAuth()
	var ne *net.Endpoint = net.NewEndpoint("test", &privateKey.PublicKey)
	na.AddTrusted("127.0.0.1", ne)

	a := net.Address{"localhost", 8000}
	n := net.NewNode(a, false, na) // pass a nil to logger pointer ~ no logging
	n.Function("/", AuthenticatedIndex, []string{"GET"}, true)

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

	if resp.Status != http.StatusUnauthorized {
		t.Error("node not rejecting unauthorized endpoints properly")
	}
}

// the auth function should default to the GlobalPermission bitmap and grant
// the request access to the route
func TestAuthGlobalPermissionPresent(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Error("Could not generate an ECDSA key pair")
	}

	var na *net.Auth = net.NewAuth()
	var ne *net.Endpoint = net.NewEndpoint("test", &privateKey.PublicKey)

	globalPermissionMap := net.NewPermission(true, false, false, false)
	ne.AddGlobalPermission(globalPermissionMap)
	na.AddTrusted("127.0.0.1", ne)

	a := net.Address{"localhost", 8000}
	n := net.NewNode(a, false, na) // pass a nil to logger pointer ~ no logging
	n.Function("/", AuthenticatedIndex, []string{"GET"}, true)

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
}

// despite an endpoint holding no global Permission bitmap, the auth function
// should default to the local Permission bitmap to determine authorization
// to the HTTP method and route
func TestAuthLocalPermissionPresent(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Error("Could not generate an ECDSA key pair")
	}

	var na *net.Auth = net.NewAuth()
	var ne *net.Endpoint = net.NewEndpoint("test", &privateKey.PublicKey)

	localPermissionMap := net.NewPermission(true, false, false, false)
	ne.AddLocalPermission("/", localPermissionMap)
	na.AddTrusted("127.0.0.1", ne)

	a := net.Address{"localhost", 8000}
	n := net.NewNode(a, false, na) // pass a nil to logger pointer ~ no logging
	n.Function("/", AuthenticatedIndex, []string{"GET"}, true)

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
}

// in the case where the user has both a global and local Permission bitmap set
// for a route, even if the global Permission bitmap denies access to an HTTP method
// the local Permission bitmap should take priority as an edge-case of permission elevation
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
	n := net.NewNode(a, false, na) // pass a nil to logger pointer ~ no logging
	n.Function("/", AuthenticatedIndex, []string{"GET", "POST"}, true)

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

	if resp.Status != http.StatusUnauthorized {
		t.Error("Node was let into a permission ")
	}

}
