package multilimiter

import (
	"sync"
	"sync/atomic"
)

//TODO: This interface may just be a waste, it's not being used right now
type Cancelable interface {
	Cancel() bool
	IsCanceled() bool
	Done() <-chan struct{}
}

var _ Cancelable = (*Canceler)(nil)

func NewCanceler() *Canceler {
	return &Canceler{
		done: make(chan struct{}),
	}
}

type Canceler struct {
	isCanceled int32
	mu         sync.Mutex
	done       chan struct{}
}

// Cancels
// returns true if we were already canceled; otherwise false
func (me *Canceler) Cancel() bool {
	me.mu.Lock()
	defer me.mu.Unlock()

	alreadyCanceled := me.isCanceled
	atomic.StoreInt32(&me.isCanceled, 1)
	if alreadyCanceled == 1 {
		return true
	}

	close(me.done)
	return false
}

func (me *Canceler) IsCanceled() bool {
	return atomic.LoadInt32(&me.isCanceled) == 1
}

func (me *Canceler) Done() <-chan struct{} {
	return me.done
}
