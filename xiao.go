// Author: CJey Hou<cjey.hou@gmail.com>

// Package xiao 提供携带名称、位置、环境变量与日志的Context，以及并发控制、序列化等常用工具
package xiao

import (
	gcontext "context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

// 进程启动时生成，并非系统的boot_id
var (
	BootID   = uuid.NewString()
	BootTime = time.Now()
)

var bootQ = SessionNameGenerator(BootID[:24])

// SessionNameGenerator 返回名称生成器，名称为prefix[0]+header+自增序号，序号不足12位时补零
func SessionNameGenerator(header string) func(...string) string {
	var counter atomic.Uint64
	return func(prefix ...string) string {
		var name = header + fmt.Sprintf("%012d", counter.Add(1))
		if len(prefix) > 0 {
			return prefix[0] + name
		}
		return name
	}
}

// SessionalContext 的名称由prefix[0]、BootID前24位与补零至12位的自增序号依次组成
func SessionalContext(prefix ...string) Context {
	return NamedContext(bootQ(prefix...))
}

// ToSessionalContext 同SessionalContext，以gctx为内部Context
func ToSessionalContext(gctx gcontext.Context, prefix ...string) Context {
	return ToNamedContext(gctx, bootQ(prefix...))
}
