package ttykit

import (
	"os"
	"sync"

	"golang.org/x/term"
)

// Attach 把本地终端切到raw模式并接到远端：键入经Escaper过滤后交给input，窗口变化交给resize，
// 行首键入~.即调用quit。返回的restore还原终端并停止转发窗口变化，可重复调用。
// 键入在单独的goroutine里读，input报错、读到尽头或断开即退出；否则一直阻塞在读上，直到进程退出
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

// IsTerminal stdin是否终端，不是即无从接管键入
func IsTerminal(stdin *os.File) bool {
	return term.IsTerminal(int(stdin.Fd()))
}
