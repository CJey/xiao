package xiao

import (
	"os"
	"sync"
	"testing"
	"time"
)

// fn panic时Wait不得返回，进程须随panic崩溃，而非正常退出
func TestLimitGroupGoPanic(t *testing.T) {
	if os.Getenv("XIAO_TEST_PANIC") != "" {
		var lg = NewLimitGroup(1)
		lg.Go(func() { panic("boom") })
		lg.Wait()
		os.Exit(0)
	}
	expectPanicExit(t)
}

// 多调用一次Done须同sync.WaitGroup一样panic，而非永久阻塞
func TestLimitGroupExtraDone(t *testing.T) {
	var recovered = make(chan any, 1)
	go func() {
		defer func() { recovered <- recover() }()
		var lg = NewLimitGroup(2)
		lg.Do()
		lg.Done()
		lg.Done()
	}()
	select {
	case x := <-recovered:
		if x == nil {
			t.Error("extra Done should panic")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("extra Done blocked")
	}
}

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

// 两个申请者同时各申请全部配额，不得各占一部分而互相等待
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
