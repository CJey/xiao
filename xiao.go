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
	var counter atomic.Uint64
	return func(prefix ...string) string {
		// return bootid related sequential name
		const h = "000000000000"
		var seq = counter.Add(1)
		var s = strconv.FormatUint(seq, 10)
		if len(prefix) > 0 {
			return prefix[0] + header + h[:12-len(s)] + s
		} else {
			return header + h[:12-len(s)] + s
		}
	}
}

// SessionalContext 的名称由BootID前24位与12位自增序号组成。
func SessionalContext(prefix ...string) Context {
	return NamedContext(bootQ(prefix...))
}

// ToSessionalContext 以给定的标准库Context作为内部基础对象，名称规则同SessionalContext。
func ToSessionalContext(gctx gcontext.Context, prefix ...string) Context {
	return ToNamedContext(gctx, bootQ(prefix...))
}
