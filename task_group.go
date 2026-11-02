package xiao

import (
	"sync"
	"sync/atomic"
)

type TaskGroup struct {
	wg  sync.WaitGroup
	cnt atomic.Int64
	evt *Event
}

func NewTaskGroup(evt *Event) *TaskGroup {
	return &TaskGroup{
		evt: evt,
	}
}

// Go 的约定同sync.WaitGroup.Go：组为空时须先于Wait调用，f panic时不放行Wait
func (tg *TaskGroup) Go(ctx Context, f func(Context)) {
	tg.cnt.Add(1)
	tg.wg.Go(func() {
		defer tg.cnt.Add(-1)
		f(ctx)
	})
}

func (tg *TaskGroup) Wait() {
	tg.wg.Wait()
}

func (tg *TaskGroup) Count() int {
	return int(tg.cnt.Load())
}

// Done 是停止信号而非任务完成信号，未关联Event时永不触发
func (tg *TaskGroup) Done() <-chan struct{} {
	if tg.evt == nil {
		return nil
	}
	return tg.evt.Yes()
}
