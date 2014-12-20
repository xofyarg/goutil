package mempool

import (
	"container/list"
	"sync"
)

type Pool struct {
	sync.Mutex
	chain *list.List
	size  int
}

func New(size int) *Pool {
	return &Pool{
		chain: list.New(),
		size:  size,
	}
}

func (p *Pool) Put(item interface{}) {
	p.Lock()
	defer p.Unlock()

	if p.chain.Len() >= p.size {
		return
	}
	p.chain.PushBack(item)
}

func (p *Pool) Get() interface{} {
	p.Lock()
	defer p.Unlock()

	if p.chain.Len() == 0 {
		return nil
	}
	return p.chain.Remove(p.chain.Front())
}
