package nonce

type Nonce uint32

type Pool struct {
	min Nonce
	max Nonce
	cur Nonce
}

func New(min, max Nonce) Pool {
	return Pool{min: min, max: max, cur: 0}
}

func (pool Pool) Next() Nonce {
	if pool.cur < pool.max {
		pool.cur++
	} else {
		pool.cur = pool.min
	}
	return pool.cur
}
