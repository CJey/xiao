package ttykit

import (
	"testing"
)

// 只认行首的~.；~~送出一个~，~后跟别的照原样送出；跨两次读也认得
func TestEscaper(t *testing.T) {
	for _, c := range []struct {
		reads []string
		sent  string
		quit  bool
	}{
		{[]string{"~."}, "", true},
		{[]string{"ls\r~.after"}, "ls\r", true},
		{[]string{"ls\n~."}, "ls\n", true},
		{[]string{"a~.b"}, "a~.b", false},
		{[]string{"~~.\r"}, "~.\r", false},
		{[]string{"~x\r~?"}, "~x\r~?", false},
		{[]string{"~\r~"}, "~\r", false},
		{[]string{"echo\r~", "."}, "echo\r", true},
		{[]string{"\x03\r~", "~", "."}, "\x03\r~.", false},
	} {
		var e = NewEscaper()
		var sent []byte
		var quit bool
		for _, r := range c.reads {
			var out []byte
			out, quit = e.Filter([]byte(r))
			if sent = append(sent, out...); quit {
				break
			}
		}
		if string(sent) != c.sent || quit != c.quit {
			t.Errorf("%q: sent %q quit %v", c.reads, sent, quit)
		}
	}
}
