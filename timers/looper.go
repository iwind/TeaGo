package timers

import "sync"

type Looper struct {
	wg         *sync.WaitGroup
	isStopping bool
}

func NewLooper() *Looper {
	return &Looper{}
}

func (this *Looper) Wait() {
	this.wg.Wait()
}

func (this *Looper) Stop() {
	this.isStopping = true
}
