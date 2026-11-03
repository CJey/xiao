package xiao

import (
	"fmt"
	"os"
	"sync/atomic"

	"go.uber.org/zap"
)

// NameJoiner 在Fork时由原名称与给定名称拼出新名称
type NameJoiner func(origin, given string) (name string)

// LocationJoiner 在Fork时由原位置与给定位置拼出新位置，field为位置在日志中的字段名
type LocationJoiner func(origin, given string) (field, location string)

// Logger 是带名称与位置的日志器，消息与数据分开传入，如Info("user created", "name", "cjey")
type Logger interface {
	// Fork 派生子日志器，名称与位置分别经NameJoiner、LocationJoiner由当前值与给定值拼出。
	// 默认以.拼接名称、以/拼接位置，给定值为空时沿用当前值；名称作为zap的logger名输出，位置默认作为字段@输出
	Fork(name, location string) Logger

	Name() string
	Location() string

	Debug(msg string, kvs ...any)
	Debugf(template string, args ...any)
	Info(msg string, kvs ...any)
	Infof(template string, args ...any)
	Warn(msg string, kvs ...any)
	Warnf(template string, args ...any)
	Error(msg string, kvs ...any)
	Errorf(template string, args ...any)
	// Panic/Fatal不受Mute影响，未配置zap时也照常panic或退出进程
	Panic(msg string, kvs ...any)
	Panicf(template string, args ...any)
	Fatal(msg string, kvs ...any)
	Fatalf(template string, args ...any)

	// With 派生附带kvs的子日志器
	With(kvs ...any) Logger
	Sync() error

	// Mute 原地屏蔽Debug至Error级别的输出，Fork时继承当时的状态
	Mute()
	Unmute()
}

type logger struct {
	zap0  *zap.SugaredLogger
	zap   *zap.SugaredLogger
	muted atomic.Bool

	nameJoiner     NameJoiner
	locationJoiner LocationJoiner

	name     string
	location string
	with     []any
}

var _ Logger = &logger{}

// defaultNameJoiner 以.拼接，任一方为空时取另一方
func defaultNameJoiner(origin, given string) string {
	const sep = "."
	if given != "" {
		if origin != "" {
			return origin + sep + given
		}
		return given
	}
	return origin
}

// defaultLocationJoiner 以/拼接，任一方为空时取另一方，字段名为@
func defaultLocationJoiner(origin, given string) (string, string) {
	const sep = "/"
	const key = "@"
	if given != "" {
		if origin != "" {
			return key, origin + sep + given
		}
		return key, given
	}
	return key, origin
}

// NewLogger 的name与location可为空；z0为nil时只有Panic/Fatal生效，nj与lj为nil时采用默认的拼接方式
func NewLogger(name, location string, z0 *zap.SugaredLogger, nj NameJoiner, lj LocationJoiner) Logger {
	return (&logger{
		zap0: z0,

		nameJoiner:     nj,
		locationJoiner: lj,
	}).fork(1, name, location)
}

func (l *logger) fork(skip int, name, location string) *logger {
	var l2 = &logger{
		zap0: l.zap0,
		zap:  l.zap0,

		nameJoiner:     l.nameJoiner,
		locationJoiner: l.locationJoiner,
	}
	l2.muted.Store(l.muted.Load())

	if skip != 0 {
		if l2.zap0 != nil {
			l2.zap0 = l2.zap0.WithOptions(zap.AddCallerSkip(skip))
		}
		if l2.zap != nil {
			l2.zap = l2.zap.WithOptions(zap.AddCallerSkip(skip))
		}
	}

	if l2.nameJoiner == nil {
		l2.name = defaultNameJoiner(l.name, name)
	} else {
		l2.name = l2.nameJoiner(l.name, name)
	}

	var field string
	if l2.locationJoiner == nil {
		field, l2.location = defaultLocationJoiner(l.location, location)
	} else {
		field, l2.location = l2.locationJoiner(l.location, location)
	}

	l2.with = make([]any, len(l.with))
	copy(l2.with, l.with)

	if l2.zap != nil {
		if l2.name != "" {
			l2.zap = l2.zap.Named(l2.name)
		}
		if l2.location != "" && field != "" {
			l2.zap = l2.zap.With(field, l2.location)
		}
		if len(l2.with) > 0 {
			l2.zap = l2.zap.With(l2.with...)
		}
	}

	return l2
}

func (l *logger) Fork(name, location string) Logger {
	return l.fork(0, name, location)
}

func (l *logger) Name() string {
	return l.name
}

func (l *logger) Location() string {
	return l.location
}

func (l *logger) Debug(msg string, kvs ...any) {
	if l.zap == nil || l.muted.Load() {
		return
	}
	l.zap.Debugw(msg, kvs...)
}

func (l *logger) Debugf(template string, args ...any) {
	if l.zap == nil || l.muted.Load() {
		return
	}
	l.zap.Debugf(template, args...)
}

func (l *logger) Info(msg string, kvs ...any) {
	if l.zap == nil || l.muted.Load() {
		return
	}
	l.zap.Infow(msg, kvs...)
}

func (l *logger) Infof(template string, args ...any) {
	if l.zap == nil || l.muted.Load() {
		return
	}
	l.zap.Infof(template, args...)
}

func (l *logger) Warn(msg string, kvs ...any) {
	if l.zap == nil || l.muted.Load() {
		return
	}
	l.zap.Warnw(msg, kvs...)
}

func (l *logger) Warnf(template string, args ...any) {
	if l.zap == nil || l.muted.Load() {
		return
	}
	l.zap.Warnf(template, args...)
}

func (l *logger) Error(msg string, kvs ...any) {
	if l.zap == nil || l.muted.Load() {
		return
	}
	l.zap.Errorw(msg, kvs...)
}

func (l *logger) Errorf(template string, args ...any) {
	if l.zap == nil || l.muted.Load() {
		return
	}
	l.zap.Errorf(template, args...)
}

func (l *logger) Panic(msg string, kvs ...any) {
	if l.zap == nil {
		panic(msg)
	}
	l.zap.Panicw(msg, kvs...)
}

func (l *logger) Panicf(template string, args ...any) {
	if l.zap == nil {
		panic(fmt.Sprintf(template, args...))
	}
	l.zap.Panicf(template, args...)
}

func (l *logger) Fatal(msg string, kvs ...any) {
	if l.zap == nil {
		exitWithoutZap(append([]any{msg}, kvs...)...)
		return
	}
	l.zap.Fatalw(msg, kvs...)
}

func (l *logger) Fatalf(template string, args ...any) {
	if l.zap == nil {
		exitWithoutZap(fmt.Sprintf(template, args...))
		return
	}
	l.zap.Fatalf(template, args...)
}

// exitWithoutZap 未配置zap时，Fatal的消息改为写入stderr，以免进程无声退出
func exitWithoutZap(a ...any) {
	fmt.Fprintln(os.Stderr, a...)
	os.Exit(1)
}

func (l *logger) With(kvs ...any) Logger {
	var l2 = l.fork(0, "", "")
	if len(kvs) > 0 {
		l2.with = append(l2.with, kvs...)
		if l2.zap != nil {
			l2.zap = l2.zap.With(kvs...)
		}
	}
	return l2
}

func (l *logger) Sync() error {
	if l.zap == nil {
		return nil
	}
	return l.zap.Sync()
}

func (l *logger) Mute() {
	l.muted.Store(true)
}

func (l *logger) Unmute() {
	l.muted.Store(false)
}
