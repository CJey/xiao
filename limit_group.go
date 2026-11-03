package xiao

import (
	"fmt"
	"runtime"
	"sync"
)

// LimitGroup 用于限制最大并发任务数
type LimitGroup interface {
	// Do 等价于 Add(1)
	Do()
	// Done 等价于 Add(-1)
	Done()
	// Go 等价于 Add(1)&fn()&Add(-1)，约定同sync.WaitGroup.Go
	Go(fn func())
	// Add 占用(tasks > 0)或释放(tasks < 0)若干任务配额，配额不足时占用操作阻塞，其余约定同sync.WaitGroup。
	// 单次数量超过最大并发数时必然死锁，故直接panic
	Add(tasks int)
	Max() uint
	// Wait 等待已占用的配额全部释放
	Wait()
}

// NewLimitGroup 的limit小于1时panic
func NewLimitGroup(limit uint) LimitGroup {
	if limit < 1 {
		panic(fmt.Errorf("LimitGroup: max[%d] less than 1", limit))
	}
	return &limitGroup{
		ch: make(chan struct{}, limit),
	}
}

// NewLimitGroupSafe 将limit修正到[1, runtime.NumCPU]之间
func NewLimitGroupSafe(limit uint) LimitGroup {
	if limit < 1 {
		limit = 1
	} else if cpu := uint(runtime.NumCPU()); cpu < limit {
		limit = cpu
	}
	return NewLimitGroup(limit)
}

func NewLimitGroupCPU() LimitGroup {
	return NewLimitGroup(uint(runtime.NumCPU()))
}

// NewLimitGroupGrace 按runtime.NumCPU适当退让，不占满全部cpu
func NewLimitGroupGrace() LimitGroup {
	var limit = uint(runtime.NumCPU())
	if limit >= 64 {
		limit /= 2
	} else if limit >= 36 {
		limit = 32
	} else if limit >= 16 {
		limit -= 4
	} else if limit >= 8 {
		limit -= 2
	} else if limit >= 4 {
		limit -= 1
	}
	return NewLimitGroup(limit)
}

type limitGroup struct {
	mu sync.Mutex // 串行化一次占用多个配额的Add，以免多个申请者各占一部分后互相等待
	wg sync.WaitGroup
	ch chan struct{}
}

func (lg *limitGroup) Max() uint {
	return uint(cap(lg.ch))
}

func (lg *limitGroup) Add(tasks int) {
	if tasks > 0 {
		if tasks > cap(lg.ch) {
			panic(fmt.Errorf("LimitGroup: Add(%d) exceeds max[%d]", tasks, cap(lg.ch)))
		}
		lg.wg.Add(tasks)
		if tasks > 1 {
			lg.mu.Lock()
			defer lg.mu.Unlock()
		}
		for range tasks {
			lg.ch <- struct{}{}
		}
	} else if tasks < 0 {
		if -tasks > cap(lg.ch) {
			panic(fmt.Errorf("LimitGroup: Add(%d) exceeds max[%d]", tasks, cap(lg.ch)))
		}
		// 先减计数：释放多于占用时同sync.WaitGroup一样panic，不致因无配额可释放而阻塞
		lg.wg.Add(tasks)
		for i := tasks; i < 0; i++ {
			<-lg.ch
		}
	}
}

func (lg *limitGroup) Do() {
	lg.Add(1)
}

func (lg *limitGroup) Done() {
	lg.Add(-1)
}

func (lg *limitGroup) Go(fn func()) {
	lg.Add(1)
	go func() {
		defer func() {
			// 同sync.WaitGroup.Go：fn panic时不释放，以免Wait提前返回，使进程先于崩溃正常退出
			if x := recover(); x != nil {
				panic(x)
			}
			lg.Done()
		}()
		fn()
	}()
}

func (lg *limitGroup) Wait() {
	lg.wg.Wait()
}
