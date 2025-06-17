package nonce

import "sync"

type Nonce uint32

type Pool struct {
	min Nonce
	max Nonce
	cur Nonce
	mux sync.Mutex
}

func New(min, max Nonce) *Pool {

	pool := new(Pool)
	pool.min = min
	pool.max = max
	pool.cur = 0

	return pool
}

func (pool *Pool) Next() Nonce {

	pool.mux.Lock()
	defer pool.mux.Unlock()

	if pool.cur < pool.max {
		pool.cur++
	} else {
		pool.cur = pool.min
	}

	return pool.cur
}
