package xiao

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestEventZeroValue(t *testing.T) {
	var evt Event
	if evt.Done() {
		t.Fatal("fresh event is done")
	}
	if !evt.Emit("a", 1) || evt.Emit("b") {
		t.Fatal("Emit should succeed only once")
	}
	select {
	case <-evt.Yes():
	default:
		t.Fatal("Yes is not closed after Emit")
	}
	if !evt.Done() || len(evt.Args) != 2 || evt.Args[0] != "a" {
		t.Fatalf("unexpected state, done=%v args=%v", evt.Done(), evt.Args)
	}
}

func TestEventConcurrentEmit(t *testing.T) {
	var evt = NewEvent()
	var wins atomic.Int32
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			<-evt.Yes()
		})
		wg.Go(func() {
			if evt.Emit() {
				wins.Add(1)
			}
		})
	}
	wg.Wait()
	if wins.Load() != 1 {
		t.Errorf("Emit succeeded %d times", wins.Load())
	}
}

func TestTaskGroupWithoutEvent(t *testing.T) {
	if NewTaskGroup(nil).Done() != nil {
		t.Error("Done without event should never fire")
	}
}
