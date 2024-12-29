package xiao

import (
	gcontext "context"
	"strconv"
	"sync/atomic"
	"time"
)

// variable alias
var (
	Canceled         = gcontext.Canceled
	DeadlineExceeded = gcontext.DeadlineExceeded
)

// type alias
type (
	CancelFunc = gcontext.CancelFunc
)

// Context extend default context package, make it better
// It should has name, location, environment and logger
type Context interface {
	// composite
	gcontext.Context

	// Fork return a copied context, if there is a new goroutine generated
	// it will use my name with a sequential number suffix started from 1
	Fork() Context
	// At return a copied context, specify the current location where it is in,
	// it should chain all locations start from root
	At(location string) Context
	ForkAt(location string) Context
	// Reborn will use gcontext.Background() instead of internal context,
	// it used for escaping internal context's cancel request
	Reborn() Context
	// RebornWith will use specified context instead of internal context,
	// it used for escaping internal context's cancel request
	RebornWith(gcontext.Context) Context
	// Name return my logger's name
	Name() string
	// Location return my logger's location
	Location() string

	// integrated base official context action
	WithCancel() (Context, CancelFunc)
	WithDeadline(time.Time) (Context, CancelFunc)
	WithTimeout(time.Duration) (Context, CancelFunc)
	WithValue(key, value any) Context

	// Env return my env
	// WARN: env value and official context value are two diffrent things
	Env() Env
	// shortcut methods of my env
	Set(key, value any)
	Get(key any) (value any, ok bool)
	GetString(key any) string
	GetInt(key any) int
	GetUint(key any) uint
	GetFloat(key any) float64
	GetBool(key any) bool

	// Logger return my logger
	Logger() Logger
	// shortcut methods of my logger
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

// Generator 定义了一个Context的生成函数，每次调用都应当返回一个新的Context
type ContextGenerator = func() Context

type context struct {
	gctx    gcontext.Context
	tracker *uint64

	env    Env
	logger Logger
}

var _ Context = (*context)(nil)

// NewContext use an official Context, an Env and a Logger to generate a new Context.
// It will use default value if not given.
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

// SimpleContext return a very simple context, without name, without location,
// and use S() as internal logger
func SimpleContext() Context {
	return NewContext(
		gcontext.Background(),
		NewEnv(),
		NewLogger("", "", _S, nil, nil),
	)
}

// ToSimpleContext 将给定的标准库Context对象作为Context的内部基础对象。
func ToSimpleContext(gctx gcontext.Context) Context {
	return NewContext(
		gctx,
		NewEnv(),
		NewLogger("", "", _S, nil, nil),
	)
}

// NamedContext 返回一个简单的Context，Context的名称可以通过参数name实现自定义。
func NamedContext(name string) Context {
	return NewContext(
		gcontext.Background(),
		NewEnv(),
		NewLogger(name, "", _S, nil, nil),
	)
}

// ToNamedContext 使用给定的标准库Context和名称列表构建一个Context。
func ToNamedContext(gctx gcontext.Context, name string) Context {
	return NewContext(
		gctx,
		NewEnv(),
		NewLogger(name, "", _S, nil, nil),
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
	return ctx.RebornWith(gcontext.Background())
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

func (ctx *context) Logger() Logger {
	if ctx.logger != nil {
		if l, yes := ctx.logger.(*logger); yes {
			return l.fork(-1, "", "")
		}
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
