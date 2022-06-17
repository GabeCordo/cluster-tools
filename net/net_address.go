package net

import "fmt"

func (a Address) String() string {
	return fmt.Sprintf("%s:%d", a.Host, a.Port)
}
