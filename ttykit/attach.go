package ttykit

import (
	"os"
	"sync"

	"golang.org/x/term"
)

// Attach 将本地终端切换至raw模式并接入远端：键入经Escaper过滤后交由input，窗口尺寸变化交由resize，
// 行首键入~.时调用quit。返回的restore恢复终端并停止转发窗口变化，可重复调用。
// 键入由独立的goroutine读取，input出错、读到EOF或断开时退出，否则持续阻塞于读取，直至进程退出
func Attach(stdin *os.File, input func([]byte) error, resize func(rows, cols int), quit func()) (func(), error) {
	var fd = int(stdin.Fd())
	var state, err = term.MakeRaw(fd)
	if err != nil {
		return nil, err
	}
	var stop = watchResize(stdin, resize)
	var restore = sync.OnceFunc(func() {
		stop()
		term.Restore(fd, state)
	})

	go func() {
		var esc = NewEscaper()
		var buf = make([]byte, 4096)
		for {
			var n, err = stdin.Read(buf)
			if n > 0 {
				var data, disconnect = esc.Filter(buf[:n])
				if len(data) > 0 && input(data) != nil {
					return
				}
				if disconnect {
					quit()
					return
				}
			}
			if err != nil {
				return
			}
		}
	}()
	return restore, nil
}

// IsTerminal 判断stdin是否为终端，非终端时无法接管键入
func IsTerminal(stdin *os.File) bool {
	return term.IsTerminal(int(stdin.Fd()))
}
