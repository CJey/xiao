package xiao

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// captureLogger 把全局zap替换为写入内存的json logger，测试结束后还原
func captureLogger(t *testing.T) *bytes.Buffer {
	var buf bytes.Buffer
	var core = zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.Lock(zapcore.AddSync(&buf)), zap.DebugLevel)
	t.Cleanup(ReplaceLogger(zap.New(core)))
	return &buf
}

func TestLogger(t *testing.T) {
	var L = NewLogger("", "", global.Load(), nil, nil)

	L = L.With("A", "B")
	L.Info("Yes")
	{
		var l = L.(*logger)
		if len(l.with) != 2 || l.with[0] != "A" || l.with[1] != "B" {
			t.Error("logger with fail")
		}
	}
	L = L.With("C", "D")
	L.Info("Yes")
	{
		var l = L.(*logger)
		if len(l.with) != 4 || l.with[0] != "A" || l.with[1] != "B" || l.with[2] != "C" || l.with[3] != "D" {
			t.Error("logger with fail")
		}
	}
}

func TestLoggerWithoutZap(t *testing.T) {
	var ctx = NewContext(nil, nil, nil)
	ctx.Logger().With("k", "v").Info("dropped")

	defer func() {
		if x := recover(); x != "must panic" {
			t.Errorf("Panic without zap: recovered %v", x)
		}
	}()
	ctx.Panic("must panic")
}

func TestPanicWhenMuted(t *testing.T) {
	var buf = captureLogger(t)
	var ctx = NamedContext("muted")
	ctx.Mute()

	var recovered any
	func() {
		defer func() { recovered = recover() }()
		ctx.Panic("panic while muted")
	}()
	if recovered == nil {
		t.Error("Panic did not panic while muted")
	}
	if !strings.Contains(buf.String(), "panic while muted") {
		t.Errorf("Panic did not log while muted: %q", buf)
	}
}

// Fatal会退出进程，在子进程中验证
func TestFatalWhenMuted(t *testing.T) {
	switch os.Getenv("XIAO_TEST_FATAL") {
	case "muted":
		var z, _ = NewSimpleLogger("info", "stderr", "console", true)
		ReplaceLogger(z)
		var ctx = NamedContext("muted")
		ctx.Mute()
		ctx.Fatal("fatal while muted")
		return
	case "nozap":
		NewContext(nil, nil, nil).Fatal("fatal without zap", "k", "v")
		return
	}

	for mode, want := range map[string]string{
		"muted": "fatal while muted",
		"nozap": "fatal without zap k v",
	} {
		var cmd = exec.Command(os.Args[0], "-test.run=^TestFatalWhenMuted$")
		cmd.Env = append(os.Environ(), "XIAO_TEST_FATAL="+mode)
		var out, err = cmd.CombinedOutput()
		var exit *exec.ExitError
		if !errors.As(err, &exit) || exit.ExitCode() != 1 {
			t.Errorf("%s: want exit code 1, got %v", mode, err)
		}
		if !strings.Contains(string(out), want) {
			t.Errorf("%s: output %q does not contain %q", mode, out, want)
		}
	}
}

// 派生时继承当时的静音状态，此后父子互不影响
func TestMuteInherit(t *testing.T) {
	var buf = captureLogger(t)
	var ctx = NamedContext("parent")
	ctx.Mute()
	var child = ctx.At("child")
	ctx.Unmute()

	ctx.Info("parent visible")
	child.Info("child hidden")
	if out := buf.String(); !strings.Contains(out, "parent visible") || strings.Contains(out, "child hidden") {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestMuteConcurrent(t *testing.T) {
	captureLogger(t)
	var ctx = NamedContext("race")
	var wg sync.WaitGroup
	wg.Go(func() {
		for range 100 {
			ctx.Info("x")
		}
	})
	wg.Go(func() {
		for range 100 {
			ctx.Mute()
			ctx.Unmute()
		}
	})
	wg.Wait()
}
