package xiao

import (
	"sync"
)

type TaskGroup struct {
	mu  sync.RWMutex
	wg  sync.WaitGroup
	evt *Event
	cnt int
}

func NewTaskGroup(evt *Event) *TaskGroup {
	return &TaskGroup{
		evt: evt,
	}
}

func (tg *TaskGroup) Go(ctx Context, f func(Context)) {
	tg.mu.Lock()
	tg.wg.Add(1)
	tg.cnt++
	tg.mu.Unlock()
	go func() {
		defer func() {
			tg.mu.Lock()
			tg.wg.Done()
			tg.cnt--
			tg.mu.Unlock()
		}()
		f(ctx)
	}()
}

func (tg *TaskGroup) Wait() {
	tg.wg.Wait()
}

func (tg *TaskGroup) Count() int {
	tg.mu.RLock()
	defer tg.mu.RUnlock()
	return tg.cnt
}

func (tg *TaskGroup) Done() <-chan struct{} {
	return tg.evt.Yes()
}
