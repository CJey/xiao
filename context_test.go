package xiao

import (
	gcontext "context"
	"regexp"
	"testing"
	"time"
)

var sessionName = regexp.MustCompile(`^` + regexp.QuoteMeta(BootID[:24]) + `\d{12}$`)

func TestSessionalContext(t *testing.T) {
	var ctx = SessionalContext().At("TestSessionalContext")
	ctx.Info("Bingo!", "username", "cjey", "sex", "male", "session", GetSession(ctx))
	if !sessionName.MatchString(ctx.Name()) {
		t.Fatalf("unexpected session name %q", ctx.Name())
	}

	ctx = ctx.At("At")
	ctx.Warn("At test")
	if ctx.Location() != "TestSessionalContext/At" {
		t.Errorf("location = %q", ctx.Location())
	}

	var ctx1 = ctx.Fork()
	ctx1.Set("abc", "xyz")
	ctx1.Warn("Fork test 1", "abc", ctx1.GetString("abc"))
	var ctx2 = ctx.Fork()
	ctx2.Warn("Fork test 2", "abc", ctx2.GetString("abc"))
	if ctx1.Name() != ctx.Name()+".1" || ctx2.Name() != ctx.Name()+".2" {
		t.Errorf("fork names = %q, %q", ctx1.Name(), ctx2.Name())
	}
	if ctx1.GetString("abc") != "xyz" || ctx2.GetString("abc") != "" || ctx.GetString("abc") != "" {
		t.Error("env should be visible only to the setter and its descendants")
	}

	var ctx11 = ctx1.Fork()
	ctx11.Warn("Fork test 1.1", "name", ctx11.Name(), "real session", GetRealSession(ctx11))
	var ctx12 = ctx1.Fork()
	var session = "7ec17674-1360-4fb1-9245-bd8d8d5866c4"
	SetSession(ctx12, session)
	ctx12.Warn("Fork test 1.2", "name", ctx12.Name(), "customized session", GetSession(ctx12))
	var ctx13 = ctx1.Fork()
	ctx13.Warn("Fork test 1.3", "name", ctx13.Name(), "session", GetSession(ctx13))
	if ctx11.Name() != ctx1.Name()+".1" || ctx11.GetString("abc") != "xyz" || GetRealSession(ctx11) != "" {
		t.Errorf("ctx11 name = %q", ctx11.Name())
	}
	if GetSession(ctx12) != session || GetSession(ctx13) != ctx13.Name() {
		t.Errorf("sessions = %q, %q", GetSession(ctx12), GetSession(ctx13))
	}
}

func TestNewNamedContext(t *testing.T) {
	var ctx = NamedContext("go test").At("TestNewNamedContext")
	ctx.Info("Bingo!", "username", "cjey", "sex", "male")
	if ctx.Name() != "go test" || ctx.Location() != "TestNewNamedContext" {
		t.Errorf("name = %q, location = %q", ctx.Name(), ctx.Location())
	}
}

func TestToSessionalContext(t *testing.T) {
	type key struct{}
	var gctx, cancel = gcontext.WithTimeout(gcontext.WithValue(gcontext.Background(), key{}, "v"), time.Minute)
	defer cancel()
	var ctx = ToSessionalContext(gctx).At("TestToContext")
	ctx.Info("Bingo!", "username", "cjey", "sex", "male")
	if _, ok := ctx.Deadline(); !ok || ctx.Value(key{}) != "v" || !sessionName.MatchString(ctx.Name()) {
		t.Errorf("should inherit the given context, name = %q", ctx.Name())
	}
}

func TestToNamedContext(t *testing.T) {
	var gctx, cancel = gcontext.WithCancel(gcontext.Background())
	var ctx = ToNamedContext(gctx, "go test").At("TestToNamedContext")
	ctx.Info("Bingo!", "username", "cjey", "sex", "male")
	cancel()
	select {
	case <-ctx.Done():
	default:
		t.Error("should follow the cancellation of the given context")
	}
}

func TestSession(t *testing.T) {
	var ctx = SessionalContext().At("TestSession")
	var session = "7ec17674-1360-4fb1-9245-bd8d8d5866c4"
	SetSession(ctx, session)
	if got := GetSession(ctx); got != session {
		t.Fatalf("GetSession = %q, want %q", got, session)
	}
}

// Reborn摆脱取消，但保留value
func TestReborn(t *testing.T) {
	type key struct{}
	var ctx, cancel = NamedContext("reborn").WithValue(key{}, "v").WithCancel()
	var reborn = ctx.Reborn()
	cancel()
	if reborn.Err() != nil || reborn.Value(key{}) != "v" || reborn.Name() != "reborn" {
		t.Errorf("err = %v, value = %v, name = %q", reborn.Err(), reborn.Value(key{}), reborn.Name())
	}
}

func TestSessionNameGenerator(t *testing.T) {
	var gen = SessionNameGenerator("h-")
	if a, b := gen(), gen("p-"); a != "h-000000000001" || b != "p-h-000000000002" {
		t.Errorf("names = %q, %q", a, b)
	}
}
