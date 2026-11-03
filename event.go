package xiao

import (
	"sync"
)

// Event 一次性事件，零值可用
type Event struct {
	once sync.Once
	mu   sync.Mutex
	done chan struct{}

	Args []any // Emit时传入的关联数据，须待事件触发后读取
}

func NewEvent() *Event {
	return new(Event)
}

func (evt *Event) ch() chan struct{} {
	evt.once.Do(func() { evt.done = make(chan struct{}) })
	return evt.done
}

// Yes 返回事件触发时关闭的chan
func (evt *Event) Yes() <-chan struct{} {
	return evt.ch()
}

// Emit 触发事件并记下args，仅首次调用生效并返回true
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

// Done 报告事件是否已触发
func (evt *Event) Done() bool {
	select {
	case <-evt.ch():
		return true
	default:
		return false
	}
}
