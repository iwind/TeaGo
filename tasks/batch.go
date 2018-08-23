package tasks

import (
	"sync"
)

type batch struct {
	tasks []func()
}

func NewBatch() *batch {
	return &batch{
		tasks: []func(){},
	}
}

func (this *batch) Add(fn func()) {
	this.tasks = append(this.tasks, fn)
}

func (this *batch) Run() {
	countTasks := len(this.tasks)
	wg := &sync.WaitGroup{}
	wg.Add(countTasks)
	for _, taskFn := range this.tasks {
		go func(taskFn func()) {
			defer wg.Add(-1)
			taskFn()
		}(taskFn)
	}
	wg.Wait()
}
