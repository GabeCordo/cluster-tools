package net

import "fmt"

const (
	GET    = "GET"
	POST   = "POST"
	PULL   = "PULL"
	DELETE = "DELETE"
)

func NewPermission(get, post, pull, delete bool) *Permission {
	perm := new(Permission)

	perm.Get = get
	perm.Post = post
	perm.Pull = pull
	perm.Delete = delete

	return perm
}

func (p Permission) Check(method string) bool {
	switch method {
	case GET:
		return p.Get
	case POST:
		return p.Post
	case DELETE:
		return p.Delete
	default:
		return p.Pull
	}
}

func (p Permission) String() string {
	return fmt.Sprintf("Permission[%t, %t, %t, %t]", p.Get, p.Delete, p.Post, p.Pull)
}

func (p Permission) Array() []string {
	var arrayRepresentation []string

	if p.Get {
		arrayRepresentation = append(arrayRepresentation, GET)
	}
	if p.Post {
		arrayRepresentation = append(arrayRepresentation, POST)
	}
	if p.Pull {
		arrayRepresentation = append(arrayRepresentation, PULL)
	}
	if p.Delete {
		arrayRepresentation = append(arrayRepresentation, DELETE)
	}

	return arrayRepresentation
}
