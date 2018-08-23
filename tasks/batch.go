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

func (batch *batch) Add(fn func()) {
	batch.tasks = append(batch.tasks, fn)
}

func (batch *batch) Run() {
	countTasks := len(batch.tasks)
	wg := &sync.WaitGroup{}
	wg.Add(countTasks)
	for _, taskFn := range batch.tasks {
		go func(taskFn func()) {
			defer wg.Add(-1)
			taskFn()
		}(taskFn)
	}
	wg.Wait()
}
