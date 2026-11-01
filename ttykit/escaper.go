// Package ttykit 把本地终端接到远端的交互式会话，同ssh -t：切raw模式，转发键入与窗口变化，行首~.断开
package ttykit

// Escaper 识别行首的~.，同ssh：Ctrl-C都送进了远端，卡住时只能靠它断开；~~送出一个~
type Escaper struct {
	lineStart bool // 处于行首
	tilde     bool // 行首收到的~，尚未送出
}

func NewEscaper() *Escaper {
	return &Escaper{lineStart: true}
}

// Filter 返回该送出的字节，以及是否要断开
func (e *Escaper) Filter(p []byte) ([]byte, bool) {
	var out = make([]byte, 0, len(p)+1)
	for _, b := range p {
		if e.tilde {
			e.tilde = false
			switch b {
			case '.':
				return out, true
			case '~':
				out = append(out, '~')
				e.lineStart = false
				continue
			}
			out = append(out, '~')
		} else if e.lineStart && b == '~' {
			e.tilde = true
			continue
		}
		out = append(out, b)
		e.lineStart = b == '\r' || b == '\n'
	}
	return out, false
}
