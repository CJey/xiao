package xiao

import (
	"sync"
	"sync/atomic"
)

// TaskGroup 等待一组任务结束，并可经关联的Event向任务发出停止信号，零值可用
type TaskGroup struct {
	wg  sync.WaitGroup
	cnt atomic.Int64
	evt *Event
}

// NewTaskGroup 的evt作为停止信号，可为nil
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

// Count 返回尚未结束的任务数
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
