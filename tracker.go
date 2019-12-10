package multilimiter

import (
	"sync"
	"sync/atomic"
	"time"
)

type ConcurrencyTracker struct {
	current int32
	total   int32
	max     int32
	locker  sync.Mutex
	started time.Time
	elapsed time.Duration
}

func (me *ConcurrencyTracker) Start() {
	atomic.StoreInt32(&me.total, 1)
	me.started = time.Now()
}

func (me *ConcurrencyTracker) Stop() {
	me.elapsed = time.Now().Sub(me.started)
}

func (me *ConcurrencyTracker) Elapsed() time.Duration {
	return me.elapsed
}

func (me *ConcurrencyTracker) Add() {
	me.locker.Lock()
	newMax := atomic.AddInt32(&me.current, 1)
	atomic.AddInt32(&me.total, 1)
	if newMax > me.max {
		me.max = newMax
	}
	me.locker.Unlock()
}

func (me *ConcurrencyTracker) Subtract() {
	me.locker.Lock()
	atomic.AddInt32(&me.current, -1)
	me.locker.Unlock()
}

func (me *ConcurrencyTracker) Current() int32 {
	return atomic.LoadInt32(&me.current)
}

func (me *ConcurrencyTracker) Total() int32 {
	return atomic.LoadInt32(&me.total)
}

func (me *ConcurrencyTracker) Max() int32 {
	return atomic.LoadInt32(&me.max)
}

func (me *ConcurrencyTracker) Rate() float64 {
	return float64(me.Total()) / me.Elapsed().Seconds()
}
