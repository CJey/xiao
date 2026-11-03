//go:build unix

package ttykit

import (
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/term"
)

// Size 返回终端窗口的行数与列数
func Size(stdin *os.File) (int, int, error) {
	var cols, rows, err = term.GetSize(int(stdin.Fd()))
	return rows, cols, err
}

func watchResize(stdin *os.File, resize func(rows, cols int)) func() {
	var winch = make(chan os.Signal, 1)
	signal.Notify(winch, syscall.SIGWINCH)
	go func() {
		for range winch {
			if rows, cols, err := Size(stdin); err == nil {
				resize(rows, cols)
			}
		}
	}()
	return func() {
		signal.Stop(winch)
		close(winch)
	}
}
