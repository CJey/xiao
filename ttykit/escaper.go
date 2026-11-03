// Package ttykit 将本地终端接入远端的交互式会话，行为同ssh -t：切换至raw模式，转发键入与窗口尺寸变化，行首键入~.断开
package ttykit

// Escaper 识别行首的~.，同ssh：raw模式下Ctrl-C亦转发至远端，会话无响应时须借此断开；~~送出单个~
type Escaper struct {
	lineStart bool // 处于行首
	tilde     bool // 行首收到的~，尚未送出
}

func NewEscaper() *Escaper {
	return &Escaper{lineStart: true}
}

// Filter 返回应转发的字节，以及是否断开会话
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
