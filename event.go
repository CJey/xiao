package xiao

import (
	"sync"
)

// Event 提供一个简单的一次性事件订阅和管理功能，零值可用
type Event struct {
	once sync.Once
	mu   sync.Mutex
	done chan struct{} // Emit触发事件后，会关闭此chan，实现事件通知效果

	Args []any // 触发事件时提供的关联数据
}

// NewEvent 初始化并返回一个*Event
func NewEvent() *Event {
	return new(Event)
}

func (evt *Event) ch() chan struct{} {
	evt.once.Do(func() { evt.done = make(chan struct{}) })
	return evt.done
}

func (evt *Event) Yes() <-chan struct{} {
	return evt.ch()
}

// 用于触发事件，支持提供可选的关联数据，只在首次触发时返回true
func (evt *Event) Emit(args ...any) bool {
	var done = evt.ch()
	evt.mu.Lock()
	defer evt.mu.Unlock()
	select {
	case <-done:
		return false
	default:
		evt.Args = args
		close(done)
		return true
	}
}

// 事件是否已经结束
func (evt *Event) Done() bool {
	select {
	case <-evt.ch():
		return true
	default:
		return false
	}
}
