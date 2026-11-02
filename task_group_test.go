package xiao

import (
	"errors"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
)

func TestTaskGroup(t *testing.T) {
	var stop = NewEvent()
	var tg = NewTaskGroup(stop)
	var ctx = NamedContext("tg")
	for range 5 {
		tg.Go(ctx, func(Context) { <-tg.Done() })
	}
	if n := tg.Count(); n != 5 {
		t.Errorf("Count = %d before stop", n)
	}
	stop.Emit()
	tg.Wait()
	if n := tg.Count(); n != 0 {
		t.Errorf("Count = %d after Wait", n)
	}
}

// 任务以Goexit结束同样计为完成
func TestTaskGroupGoexit(t *testing.T) {
	var tg = NewTaskGroup(nil)
	tg.Go(NamedContext("tg"), func(Context) { runtime.Goexit() })
	tg.Wait()
	if n := tg.Count(); n != 0 {
		t.Errorf("Count = %d after Wait", n)
	}
}

func TestTaskGroupWithoutEvent(t *testing.T) {
	if NewTaskGroup(nil).Done() != nil {
		t.Error("Done without event should never fire")
	}
}

// 任务panic时Wait不得返回，进程须随panic崩溃，而非正常退出
func TestTaskGroupGoPanic(t *testing.T) {
	if os.Getenv("XIAO_TEST_PANIC") != "" {
		var tg = NewTaskGroup(nil)
		tg.Go(NamedContext("tg"), func(Context) { panic("boom") })
		tg.Wait()
		os.Exit(0)
	}
	expectPanicExit(t)
}

// expectPanicExit 在子进程中重跑当前测试，断言进程以panic("boom")崩溃告终
func expectPanicExit(t *testing.T) {
	var cmd = exec.Command(os.Args[0], "-test.run=^"+t.Name()+"$")
	cmd.Env = append(os.Environ(), "XIAO_TEST_PANIC=1")
	var out, err = cmd.CombinedOutput()
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 2 || !strings.Contains(string(out), "boom") {
		t.Errorf("want crash by panic, got %v: %s", err, out)
	}
}
