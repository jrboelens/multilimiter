package multilimiter

import (
	"sync"
	"sync/atomic"
)

//TODO: This interface may just be a waste, it's not being used right now
type Cancelable interface {
	Cancel(funcs ...func()) bool
	IsCanceled() bool
}

var _ Cancelable = (*Canceler)(nil)

type Canceler struct {
	isCanceled int32
	mu         sync.Mutex
}

// Cancels
// returns true if we were already canceled; otherwise false
// executes fn's if the state changes from not canceled to canceled
func (me *Canceler) Cancel(funcs ...func()) bool {
	me.mu.Lock()
	defer me.mu.Unlock()

	alreadyCanceled := me.isCanceled
	me.isCanceled = 1
	if alreadyCanceled == 1 {
		return true
	}

	for _, fn := range funcs {
		fn()
	}

	return false
}

func (me *Canceler) IsCanceled() bool {
	return atomic.LoadInt32(&me.isCanceled) == 1
}
