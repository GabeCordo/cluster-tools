package net

type NodeIllegalActionError struct{}

func (e *NodeIllegalActionError) Error() string {
	return "Illegal Action Given the Node's Current State"
}
