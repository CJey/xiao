//go:build !unix

package ttykit

import (
	"os"
	"time"

	"golang.org/x/term"
)

// 没有SIGWINCH，只能轮询
const resizePoll = 500 * time.Millisecond

// Size 终端窗口的行数与列数。控制台的窗口大小要经输出句柄取，stdin的句柄取不到
func Size(_ *os.File) (int, int, error) {
	var cols, rows, err = term.GetSize(int(os.Stdout.Fd()))
	return rows, cols, err
}

func watchResize(stdin *os.File, resize func(rows, cols int)) func() {
	var done = make(chan struct{})
	go func() {
		var ticker = time.NewTicker(resizePoll)
		defer ticker.Stop()
		var rows, cols, _ = Size(stdin)
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
			}
			if r, c, err := Size(stdin); err == nil && (r != rows || c != cols) {
				rows, cols = r, c
				resize(rows, cols)
			}
		}
	}()
	return func() { close(done) }
}
