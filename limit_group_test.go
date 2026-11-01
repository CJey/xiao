package xiao

import (
	"sync"
	"testing"
	"time"
)

func TestLimitGroup(t *testing.T) {
	var lg = NewLimitGroup(3)
	var mu sync.Mutex
	var running, peak, done int
	for range 50 {
		lg.Go(func() {
			mu.Lock()
			running++
			peak = max(peak, running)
			mu.Unlock()

			time.Sleep(time.Millisecond)

			mu.Lock()
			running--
			done++
			mu.Unlock()
		})
	}
	lg.Wait()
	if done != 50 || peak > 3 {
		t.Errorf("done=%d peak=%d", done, peak)
	}
}

// 两个申请者同时各要全部配额，不得各占一部分而互相等待
func TestLimitGroupMultiAdd(t *testing.T) {
	var lg = NewLimitGroup(2)
	var finished = make(chan struct{})
	go func() {
		var wg sync.WaitGroup
		for range 2 {
			wg.Go(func() {
				for range 1000 {
					lg.Add(2)
					lg.Add(-2)
				}
			})
		}
		wg.Wait()
		close(finished)
	}()
	select {
	case <-finished:
	case <-time.After(10 * time.Second):
		t.Fatal("deadlocked")
	}
}

func TestLimitGroupOverflow(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("Add beyond max should panic")
		}
	}()
	NewLimitGroup(2).Add(3)
}
