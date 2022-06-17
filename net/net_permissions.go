package net

import "fmt"

func NewPermission(get, post, pull, delete bool) Permission {
	return Permission{get, post, pull, delete}
}

func (p Permission) Check(method string) bool {
	switch method {
	case "get":
		return p.Get
	case "post":
		return p.Post
	case "delete":
		return p.Delete
	default:
		return p.Pull
	}
}

func (p Permission) String() string {
	return fmt.Sprintf("Permission[%t, %t, %t, %t]", p.Get, p.Delete, p.Post, p.Pull)
}
