package xiao

import (
	"net"
	"reflect"
	"sync"
	"time"
)

// Env should used as a map, but it has an inherited mode, overlay liked.
// If you set, the value should be store at local.
// If you get, the value should be get from local first, otherwise from parent
type Env interface {
	// Fork return an inherited sub Env, and I am it's parent
	Fork() Env

	// Set always set key & value at local storage
	Set(key, value any)
	// Get always check local, if the key not exists, then check parent
	Get(key any) (value any, ok bool)
	Has(key any) (ok bool)
	Keys() []any

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

// NewEnv return a simple Env, use sync.Map as it's storage
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
	// from local
	if value, ok := e.vals.Load(key); ok {
		return value, ok
	}
	// otherwise from parent
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

func (e *env) GetInt(key any) int {
	if value, ok := e.Get(key); ok {
		return value.(int)
	}
	return 0
}

func (e *env) GetInt64(key any) int64 {
	if value, ok := e.Get(key); ok {
		return value.(int64)
	}
	return 0
}

func (e *env) GetUint(key any) uint {
	if value, ok := e.Get(key); ok {
		return value.(uint)
	}
	return 0
}

func (e *env) GetUint64(key any) uint64 {
	if value, ok := e.Get(key); ok {
		return value.(uint64)
	}
	return 0
}

func (e *env) GetBool(key any) bool {
	if value, ok := e.Get(key); ok {
		return value.(bool)
	}
	return false
}

func (e *env) GetFloat(key any) float64 {
	if value, ok := e.Get(key); ok {
		return value.(float64)
	}
	return 0.0
}

func (e *env) GetString(key any) string {
	if value, ok := e.Get(key); ok {
		return value.(string)
	}
	return ""
}

func (e *env) GetIP(key any) net.IP {
	if value, ok := e.Get(key); ok {
		return value.(net.IP)
	}
	return nil
}

func (e *env) GetAddr(key any) net.Addr {
	if value, ok := e.Get(key); ok {
		return value.(net.Addr)
	}
	return nil
}

func (e *env) GetTime(key any) time.Time {
	if value, ok := e.Get(key); ok {
		return value.(time.Time)
	}
	return time.Time{}
}

func (e *env) GetDuration(key any) time.Duration {
	if value, ok := e.Get(key); ok {
		return value.(time.Duration)
	}
	return 0
}
