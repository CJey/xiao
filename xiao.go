// Author: CJey Hou<cjey.hou@gmail.com>
package xiao

import (
	gcontext "context"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

var (
	BootID   = uuid.NewString()
	BootTime = time.Now()
)

// 基于BootID来做Session前缀
var bootQ = SessionNameGenerator(BootID[:24])

// SessionNameGenerator 返回一个命名生成器，会自动在header后面追加12位长的自增字符串，
// 同时支持提供额外的前缀，会被附加在header之前。
func SessionNameGenerator(header string) func(...string) string {
	var counter uint64
	return func(prefix ...string) string {
		// return bootid related sequential name
		const h = "000000000000"
		var seq = atomic.AddUint64(&counter, 1)
		var s = strconv.FormatUint(seq, 10)
		if len(prefix) > 0 {
			return prefix[0] + header + h[:12-len(s)] + s
		} else {
			return header + h[:12-len(s)] + s
		}
	}
}

// SessinalContext 返回一个简单的Context，命名部分使用自动生成的uuid。
func SessionalContext(prefix ...string) Context {
	return NamedContext(bootQ(prefix...))
}

// ToSesionalContext 将给定的标准库Context对象对位Context的内部基础对象，并自动生成uuid作为name。
func ToSessionalContext(gctx gcontext.Context, prefix ...string) Context {
	return ToNamedContext(gctx, bootQ(prefix...))
}
