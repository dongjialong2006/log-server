package server

import (
	"sync"
	"sync/atomic"
)

type option struct {
	sync.RWMutex
	ttl       int64
	addr      string
	stop      string
	timeout   string
	identity  string
	namespace string
}

func (o *option) UpdateTTL(ttl int64) {
	o.Lock()
	defer o.Unlock()
	o.ttl = ttl
}

func (o *option) GetTTL() int64 {
	o.RLock()
	defer o.RUnlock()
	return o.ttl
}

func (o *option) UpdateStop(stop string) {
	o.Lock()
	defer o.Unlock()
	o.stop = stop
}

func (o *option) GetStop() string {
	o.RLock()
	defer o.RUnlock()
	return o.stop
}

func (o *option) UpdateIdentity(identity string) {
	o.Lock()
	defer o.Unlock()
	o.identity = identity
}

func (o *option) GetIdentity() string {
	o.RLock()
	defer o.RUnlock()
	return o.identity
}

func (o *option) UpdateTimeout(timeout string) {
	o.Lock()
	defer o.Unlock()
	o.timeout = timeout
}

func (o *option) GetTimeout() string {
	o.RLock()
	defer o.RUnlock()
	return o.timeout
}

type attr struct {
	end  int32
	num  int64
	size int64
}

func (a *attr) update(num int64) {
	atomic.StoreInt64(&a.num, num)
}

func (a *attr) getNum() int64 {
	return atomic.LoadInt64(&a.num)
}
