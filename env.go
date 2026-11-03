package xiao

import (
	"net"
	"reflect"
	"sync"
	"time"
)

// Env 是分层的键值存储，Fork派生的下层可遮盖上层的同名键
type Env interface {
	Fork() Env

	// Set 只写本层，key为nil或不可比较时panic
	Set(key, value any)
	// Get 先查本层，未命中再逐级查上层；上层在派生之后写入的值同样可见
	Get(key any) (value any, ok bool)
	Has(key any) (ok bool)
	// Keys 返回本层及各上层的全部键
	Keys() []any

	// GetXxx 键不存在或值为nil时返回零值，类型不符时panic
	GetInt(key any) int
	GetInt64(key any) int64
	GetUint(key any) uint
	GetUint64(key any) uint64
	GetBool(key any) bool
	GetFloat(key any) float64
	GetString(key any) string
	GetIP(key any) net.IP
	GetAddr(key any) net.Addr
	GetTime(key any) time.Time
	GetDuration(key any) time.Duration
}

type env struct {
	parent *env

	vals sync.Map
}

var _ Env = &env{}

func NewEnv() Env {
	return &env{}
}

func (e *env) fork() *env {
	return &env{
		parent: e,
	}
}

func (e *env) Fork() Env {
	return e.fork()
}

func (e *env) Set(key, value any) {
	if key == nil {
		panic("nil key")
	}
	if !reflect.TypeOf(key).Comparable() {
		panic("key is not comparable")
	}
	e.vals.Store(key, value)
}

func (e *env) Get(key any) (value any, ok bool) {
	if value, ok := e.vals.Load(key); ok {
		return value, ok
	}
	if e.parent != nil {
		return e.parent.Get(key)
	}
	return nil, false
}

func (e *env) Has(key any) (ok bool) {
	_, ok = e.Get(key)
	return
}

func (e *env) keys() map[any]struct{} {
	var keys map[any]struct{}
	if e.parent != nil {
		keys = e.parent.keys()
	} else {
		keys = make(map[any]struct{})
	}
	e.vals.Range(func(k, v any) bool {
		keys[k] = struct{}{}
		return true
	})
	return keys
}

func (e *env) Keys() []any {
	var (
		idx  = e.keys()
		keys = make([]any, 0, len(idx))
	)
	for k := range idx {
		keys = append(keys, k)
	}
	return keys
}

func getAs[T any](e *env, key any) T {
	if value, ok := e.Get(key); ok && value != nil {
		return value.(T)
	}
	var zero T
	return zero
}

func (e *env) GetInt(key any) int                { return getAs[int](e, key) }
func (e *env) GetInt64(key any) int64            { return getAs[int64](e, key) }
func (e *env) GetUint(key any) uint              { return getAs[uint](e, key) }
func (e *env) GetUint64(key any) uint64          { return getAs[uint64](e, key) }
func (e *env) GetBool(key any) bool              { return getAs[bool](e, key) }
func (e *env) GetFloat(key any) float64          { return getAs[float64](e, key) }
func (e *env) GetString(key any) string          { return getAs[string](e, key) }
func (e *env) GetIP(key any) net.IP              { return getAs[net.IP](e, key) }
func (e *env) GetAddr(key any) net.Addr          { return getAs[net.Addr](e, key) }
func (e *env) GetTime(key any) time.Time         { return getAs[time.Time](e, key) }
func (e *env) GetDuration(key any) time.Duration { return getAs[time.Duration](e, key) }
