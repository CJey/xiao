# Xiao

My personal basic golang package

[![Go Reference](https://pkg.go.dev/badge/github.com/cjey/xiao.svg)](https://pkg.go.dev/github.com/cjey/xiao)

```go
import "github.com/cjey/xiao"
```

## Context

Context在实现标准库Context接口的基础上，附带了环境变量与日志能力。

### 关键特性

**兼容**

实现了标准库的Context接口，可用于任何接受context.Context的场合。

**WithX**

即WithCancel、WithDeadline、WithTimeout与WithValue，用于便捷地派生Context。

**Fork**

派生出新的一层环境变量，并在名称之后追加自增的数字后缀，位置保持不变。

典型场景：一次请求中需要并发执行多个子任务时，各子任务使用Fork得到的Context，日志中即可据名称区分各子任务。

派生的Context用于异步任务时须谨慎：源Context通常在同步请求结束时即被取消，异步任务会随之中断。此时应改用Reborn派生，它摆脱源Context的取消与截止时间，同时保留名称、位置、环境变量与value。

**At**

派生出新的一层环境变量，并在原有位置之后追加新的位置名称，名称保持不变。位置随派生逐级继承，并在日志输出时自动携带，从而在日志中呈现逻辑上的调用层次。

此能力宜用在关键的逻辑节点上，不建议在每个函数中使用，否则与每条日志都打印一次调用栈无异。

**Set/Get**

Env的快捷方法，提供类似环境变量的能力。每次派生Context，都会同步派生出新的一层变量空间。

Set只写入当前层，不影响上层；Get优先查找当前层，未找到时逐级向上层查找。

**Debug/Info/...**

Logger的快捷方法，提供基本的日志能力。内部使用zap.SugaredLogger，日志格式经过重新调整；如需自定义格式，可自行替换全局logger。

只需在默认风格上调整日志级别、输出位置或编码时，使用UseSimpleLogger即可；已有现成的zap.Logger时，则使用ReplaceLogger。替换只影响此后创建的Context。

默认：UseSimpleLogger("info", "stderr", "console", false)

info+json：UseSimpleLogger("info", "stderr", "json", false)

**Mute/Unmute**

原地屏蔽或恢复当前Context的Debug至Error级日志，适用于高频请求按需静默；派生出的Context继承派生时的状态。Panic/Fatal不受影响，照常输出并终止。

### NamedContext

Context支持自定义名称，一般用于伴随服务生命周期的任务，或定时执行的任务。

须注意，Context创建时其内部的logger即已固定。对于定时任务，如需在运行时调整logger，应在每次执行时重新创建同名的Context。

```go
var ctx = xiao.NamedContext("ZookeeperMonitor")
ctx.Info("started")
// console encode
// 2020-04-11 21:24:45.361  info    ZookeeperMonitor  app/main.go:12  started
// json encode
// {"L":"info","T":"2020-04-11 21:24:45.361","N":"ZookeeperMonitor","C":"app/main.go:12","M":"started"}

ctx = ctx.At("GetID")
ctx.Set("id", "123")
ctx.Info("got it!", "id", ctx.GetString("id"))
// console encode
// 2020-04-11 21:24:45.361  info    ZookeeperMonitor  app/main.go:16  got it!  {"@": "GetID", "id": "123"}
// json encode
// {"L":"info","T":"2020-04-11 21:24:45.361","N":"ZookeeperMonitor","C":"app/main.go:16","M":"got it!","@":"GetID","id":"123"}
```

### SessionalContext

也可以使用自动生成uuid风格名称的Context，适用于在请求开始时为本次请求创建一个SessionalContext。

生成的名称形似uuid，由BootID的前24位与12位十进制自增序号组成。

如需定制命名规则，可用SessionNameGenerator创建自己的生成器，再交由NamedContext使用。

```go
var ctx = xiao.SessionalContext()
ctx.Info("started")
// console encode
// 2020-04-11 21:24:45.361  info    9b2119d3-7f37-4033-8c19-000000000001  app/main.go:12  started
// json encode
// {"L":"info","T":"2020-04-11 21:24:45.361","N":"9b2119d3-7f37-4033-8c19-000000000001","C":"app/main.go:12","M":"started"}

ctx = ctx.ForkAt("AsyncGetID")
ctx.Info("show default session", "session", xiao.GetSession(ctx))
// console encode
// 2020-04-11 21:24:45.361  info    9b2119d3-7f37-4033-8c19-000000000001.1  app/main.go:15  show default session  {"@": "AsyncGetID", "session": "9b2119d3-7f37-4033-8c19-000000000001.1"}
// json encode
// {"L":"info","T":"2020-04-11 21:24:45.361","N":"9b2119d3-7f37-4033-8c19-000000000001.1","C":"app/main.go:15","M":"show default session","@":"AsyncGetID","session":"9b2119d3-7f37-4033-8c19-000000000001.1"}

xiao.SetSession(ctx, "7ec17674-1360-4fb1-9245-bd8d8d5866c4")
ctx.Info("use my own session", "session", xiao.GetSession(ctx))
// console encode
// 2020-04-11 21:24:45.361  info    9b2119d3-7f37-4033-8c19-000000000001.1  app/main.go:19  use my own session  {"@": "AsyncGetID", "session": "7ec17674-1360-4fb1-9245-bd8d8d5866c4"}
// json encode
// {"L":"info","T":"2020-04-11 21:24:45.361","N":"9b2119d3-7f37-4033-8c19-000000000001.1","C":"app/main.go:19","M":"use my own session","@":"AsyncGetID","session":"7ec17674-1360-4fb1-9245-bd8d8d5866c4"}
```

### ToContext

NamedContext与SessionalContext均以标准库的Background作为内部Context。如需使用自定义的Context，或转换标准库的Context，可使用**ToSessionalContext**与**ToNamedContext**。

## 其他

- **Event / TaskGroup / LimitGroup**：一次性事件通知、带停止信号的任务组、并发配额限制
- **Json / RawJson / JsonPretty / Gzip / Gunzip**：常用的序列化与压缩工具，`Json`的结果可直接作为日志字段输出
- **gerror**：将业务错误码编入grpc的带外错误，handler可直接返回`*GError`，调用方以`Decode`还原
- **netkit**：地址补全与去重、本机管理网与公网地址探测、TCPing与ICMPing
- **ttykit**：将本地终端接入远端的交互式会话，同`ssh -t`
