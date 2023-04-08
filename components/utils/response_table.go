package utils

import "sync"

type ResponseTable struct {
	responses map[uint32]any

	mutex sync.RWMutex
}

func NewResponseTable() *ResponseTable {
	table := new(ResponseTable)
	table.responses = make(map[uint32]any)
	return table
}

func (responseTable *ResponseTable) Write(nonce uint32, response any) {
	responseTable.mutex.Lock()
	defer responseTable.mutex.Unlock()

	responseTable.responses[nonce] = response
}

func (responseTable *ResponseTable) Lookup(nonce uint32) (response any, found bool) {
	responseTable.mutex.RLock()
	defer responseTable.mutex.RUnlock()

	if response, found := responseTable.responses[nonce]; found {
		return response, found
	} else {
		return nil, found
	}
}
