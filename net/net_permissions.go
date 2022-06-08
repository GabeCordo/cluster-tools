package net

func NewPermission(get, post, pull, delete bool) Permission {
	return Permission{get, post, pull, delete}
}

func (p Permission) Check(method string) bool {
	switch method {
	case "get":
		return p.get
	case "post":
		return p.post
	case "delete":
		return p.delete
	default:
		return p.pull
	}
}
