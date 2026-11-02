package netkit

import (
	"fmt"
	"net"
	"strings"
	"testing"
)

func TestUniqCSVAddrs(t *testing.T) {
	var addrs = UniqCSVAddrs(2181, "127.0.0.1", "", "127.0.0.2:111", "127.0.0.3:2181,127.0.0.4:2181", "中:文")
	t.Log("addrs " + strings.Join(addrs, ","))

	var noport = TrimPort(2181, "127.0.0.1:2181")
	fmt.Printf("noport %#v\n", noport)

	var str = PrettyCSVAddr(2181, "127.0.0.1:2181")
	t.Log("str " + str)
}

// 端口为0或留空、IPv6带方括号时的补全
func TestUniqCSVAddrsEdges(t *testing.T) {
	for _, c := range []struct {
		port uint16
		in   []string
		want string
	}{
		{2181, []string{"[::1]:0", "127.0.0.1:0"}, "[::1]:2181,127.0.0.1:2181"},
		{2181, []string{"[::1]", "127.0.0.1:", "[::1]:"}, "[::1]:2181,127.0.0.1:2181"},
		{0, []string{"[::1]", "127.0.0.1:", ":0", "a:80"}, "::1,127.0.0.1,a:80"},
	} {
		if got := strings.Join(UniqCSVAddrs(c.port, c.in...), ","); got != c.want {
			t.Errorf("UniqCSVAddrs(%d, %q) = %s, want %s", c.port, c.in, got, c.want)
		}
	}
}

func TestAddrCompletion(t *testing.T) {
	var ip = net.ParseIP("10.0.0.1")
	for raw, want := range map[string]string{
		"":           "10.0.0.1:80",
		"8080":       "10.0.0.1:8080",
		":8080":      "10.0.0.1:8080",
		"1.1.1.1":    "1.1.1.1:80",
		"1.1.1.1:":   "1.1.1.1:80",
		"::1":        "[::1]:80",
		"[::1]":      "[::1]:80",
		"[::1]:":     "[::1]:80",
		"[::1]:8080": "[::1]:8080",
	} {
		if got, err := AddrCompletion(raw, ip, 80); err != nil || got != want {
			t.Errorf("AddrCompletion(%q) = %s, %v, want %s", raw, got, err, want)
		}
	}
}
