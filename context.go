package xiao

import (
	gcontext "context"
	"strconv"
	"sync/atomic"
	"time"
)

var (
	Canceled         = gcontext.Canceled
	DeadlineExceeded = gcontext.DeadlineExceeded
)

type (
	CancelFunc = gcontext.CancelFunc
)

// Context 在标准库Context之外附带名称、位置、环境变量与日志，每次派生都为环境变量另起一层
type Context interface {
	gcontext.Context

	// Fork 派生子Context，名称追加从1起的自增序号，用于区分并发的子任务
	Fork() Context
	// At 派生子Context，在当前位置之后追加location，名称不变
	At(location string) Context
	// ForkAt 同时执行Fork与At
	ForkAt(location string) Context
	// Reborn 摆脱内部Context的取消与截止时间，保留其中的value
	Reborn() Context
	// RebornWith 以gctx替换内部Context，为nil时用Background
	RebornWith(gcontext.Context) Context
	Name() string
	Location() string

	WithCancel() (Context, CancelFunc)
	WithDeadline(time.Time) (Context, CancelFunc)
	WithTimeout(time.Duration) (Context, CancelFunc)
	WithValue(key, value any) Context

	// Env 返回环境变量，与标准库Context的value互不相通
	Env() Env
	// Env的快捷方法
	Set(key, value any)
	Get(key any) (value any, ok bool)
	GetString(key any) string
	GetInt(key any) int
	GetUint(key any) uint
	GetFloat(key any) float64
	GetBool(key any) bool

	// Mute 原地屏蔽本Context的Debug至Error级日志，派生时继承当时的状态；Panic/Fatal不受影响
	Mute()
	Unmute()
	Logger() Logger
	// Logger的快捷方法
	Debug(msg string, kvs ...any)
	Debugf(template string, args ...any)
	Info(msg string, kvs ...any)
	Infof(template string, args ...any)
	Warn(msg string, kvs ...any)
	Warnf(template string, args ...any)
	Error(msg string, kvs ...any)
	Errorf(template string, args ...any)
	Panic(msg string, kvs ...any)
	Panicf(template string, args ...any)
	Fatal(msg string, kvs ...any)
	Fatalf(template string, args ...any)
}

// ContextGenerator 每次调用都应返回新的Context
type ContextGenerator = func() Context

type context struct {
	gctx    gcontext.Context
	tracker *uint64

	env    Env
	logger Logger
}

var _ Context = (*context)(nil)

// NewContext 以给定的标准库Context、Env与Logger组装Context，gctx与env为nil时取默认值；
// log为nil时日志一律丢弃，Panic/Fatal除外
func NewContext(gctx gcontext.Context, env Env, log Logger) Context {
	if gctx == nil {
		gctx = gcontext.Background()
	}
	if env == nil {
		env = NewEnv()
	}
	if log == nil {
		log = NewLogger("", "", nil, nil, nil)
	}

	if l, yes := log.(*logger); yes {
		log = l.fork(1, "", "")
	}

	var tracker uint64
	return &context{
		gctx:    gctx,
		tracker: &tracker,

		env:    env,
		logger: log,
	}
}

// SimpleContext 返回无名称、无位置、使用全局logger的Context
func SimpleContext() Context {
	return ToNamedContext(gcontext.Background(), "")
}

// ToSimpleContext 同SimpleContext，以gctx为内部Context
func ToSimpleContext(gctx gcontext.Context) Context {
	return ToNamedContext(gctx, "")
}

// NamedContext 同SimpleContext，名称为name
func NamedContext(name string) Context {
	return ToNamedContext(gcontext.Background(), name)
}

// ToNamedContext 同NamedContext，以gctx为内部Context
func ToNamedContext(gctx gcontext.Context, name string) Context {
	return NewContext(
		gctx,
		NewEnv(),
		NewLogger(name, "", global.Load(), nil, nil),
	)
}

func (ctx *context) fork(name, location string) *context {
	return &context{
		gctx:    ctx.gctx,
		tracker: ctx.tracker,

		env:    ctx.env.Fork(),
		logger: ctx.logger.Fork(name, location),
	}
}

func (ctx *context) Deadline() (deadline time.Time, ok bool) {
	return ctx.gctx.Deadline()
}

func (ctx *context) Done() <-chan struct{} {
	return ctx.gctx.Done()
}

func (ctx *context) Err() error {
	return ctx.gctx.Err()
}

func (ctx *context) Value(key any) any {
	return ctx.gctx.Value(key)
}

func (ctx *context) Fork() Context {
	return ctx.ForkAt("")
}

func (ctx *context) At(location string) Context {
	return ctx.fork("", location)
}

func (ctx *context) ForkAt(location string) Context {
	var seq = atomic.AddUint64(ctx.tracker, 1)
	var newctx = ctx.fork(strconv.FormatUint(seq, 10), location)
	var tracker uint64
	newctx.tracker = &tracker
	return newctx
}

func (ctx *context) Reborn() Context {
	return ctx.RebornWith(gcontext.WithoutCancel(ctx.gctx))
}

func (ctx *context) RebornWith(gctx gcontext.Context) Context {
	if gctx == nil {
		gctx = gcontext.Background()
	}
	var newctx = ctx.fork("", "")
	newctx.gctx = gctx
	return newctx
}

func (ctx *context) Name() string {
	return ctx.logger.Name()
}

func (ctx *context) Location() string {
	return ctx.logger.Location()
}

func (ctx *context) WithCancel() (Context, CancelFunc) {
	var newctx = ctx.fork("", "")
	var newgctx, f = gcontext.WithCancel(newctx.gctx)
	newctx.gctx = newgctx
	return newctx, f
}

func (ctx *context) WithDeadline(d time.Time) (Context, CancelFunc) {
	var newctx = ctx.fork("", "")
	var newgctx, f = gcontext.WithDeadline(newctx.gctx, d)
	newctx.gctx = newgctx
	return newctx, f
}

func (ctx *context) WithTimeout(timeout time.Duration) (Context, CancelFunc) {
	var newctx = ctx.fork("", "")
	var newgctx, f = gcontext.WithTimeout(newctx.gctx, timeout)
	newctx.gctx = newgctx
	return newctx, f
}

func (ctx *context) WithValue(key, value any) Context {
	var newctx = ctx.fork("", "")
	newctx.gctx = gcontext.WithValue(newctx.gctx, key, value)
	return newctx
}

func (ctx *context) Env() Env {
	return ctx.env
}

func (ctx *context) Mute() {
	ctx.logger.Mute()
}
func (ctx *context) Unmute() {
	ctx.logger.Unmute()
}
func (ctx *context) Logger() Logger {
	if l, yes := ctx.logger.(*logger); yes {
		return l.fork(-1, "", "")
	}
	return ctx.logger
}

func (ctx *context) Set(key, value any) {
	ctx.env.Set(key, value)
}

func (ctx *context) Get(key any) (value any, ok bool) {
	return ctx.env.Get(key)
}

func (ctx *context) GetString(key any) string {
	return ctx.env.GetString(key)
}

func (ctx *context) GetInt(key any) int {
	return ctx.env.GetInt(key)
}

func (ctx *context) GetUint(key any) uint {
	return ctx.env.GetUint(key)
}

func (ctx *context) GetFloat(key any) float64 {
	return ctx.env.GetFloat(key)
}

func (ctx *context) GetBool(key any) bool {
	return ctx.env.GetBool(key)
}

func (ctx *context) Debug(msg string, kvs ...any) {
	ctx.logger.Debug(msg, kvs...)
}
func (ctx *context) Debugf(template string, args ...any) {
	ctx.logger.Debugf(template, args...)
}

func (ctx *context) Info(msg string, kvs ...any) {
	ctx.logger.Info(msg, kvs...)
}
func (ctx *context) Infof(template string, args ...any) {
	ctx.logger.Infof(template, args...)
}

func (ctx *context) Warn(msg string, kvs ...any) {
	ctx.logger.Warn(msg, kvs...)
}
func (ctx *context) Warnf(template string, args ...any) {
	ctx.logger.Warnf(template, args...)
}

func (ctx *context) Error(msg string, kvs ...any) {
	ctx.logger.Error(msg, kvs...)
}
func (ctx *context) Errorf(template string, args ...any) {
	ctx.logger.Errorf(template, args...)
}

func (ctx *context) Panic(msg string, kvs ...any) {
	ctx.logger.Panic(msg, kvs...)
}
func (ctx *context) Panicf(template string, args ...any) {
	ctx.logger.Panicf(template, args...)
}

func (ctx *context) Fatal(msg string, kvs ...any) {
	ctx.logger.Fatal(msg, kvs...)
}
func (ctx *context) Fatalf(template string, args ...any) {
	ctx.logger.Fatalf(template, args...)
}
