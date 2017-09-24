package mgop

import (
	"sync"
	_"fmt"
)

type upstreamSessionPool interface {
	appendSession(*SessionWrapper)
	getBest() *SessionWrapper
	foreach(func(sw *SessionWrapper), bool)
}

/**
one kind of upstream for polling, just for test
suppose support for insert,update and findAndModify is OK
 */

type pollingSessionPool struct {
	wrappers []*SessionWrapper
	maxSize  int
	next     int
	mutex    sync.RWMutex
}

func (p *pollingSessionPool)getBest() *SessionWrapper {
	p.mutex.Lock()
	p.next = (p.next + 1) % len(p.wrappers)
	cur := p.next
	p.mutex.Unlock()
	//fmt.Printf("get %d\n", cur)
	p.wrappers[cur].atomicAcquire()
	return p.wrappers[cur]
}

func (p *pollingSessionPool)foreach(eachFunc func(sw *SessionWrapper), readonly bool) {
	if readonly {
		p.mutex.RLock()
		defer p.mutex.RUnlock()
	} else {
		p.mutex.Lock()
		defer p.mutex.Unlock()
	}
	for _, s := range p.wrappers {
		eachFunc(s)
	}
}

func newPollingSessionPool(maxSize int) upstreamSessionPool {
	p := &pollingSessionPool{
		maxSize:maxSize,
		next:-1,
	}
	p.wrappers = make([]*SessionWrapper, 0, maxSize)
	return p
}

func (p *pollingSessionPool)appendSession(sw *SessionWrapper) {
	p.mutex.Lock()
	p.wrappers = append(p.wrappers, sw)
	p.mutex.Unlock()
}

func (p *pollingSessionPool)size() int {
	p.mutex.RLock()
	p.mutex.RUnlock()
	return len(p.wrappers)
}
