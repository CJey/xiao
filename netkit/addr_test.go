package netkit

import (
	"net"
	"slices"
	"strings"
	"testing"
)

// 补全缺失的端口，剔除端口非法的地址
func TestUniqCSVAddrs(t *testing.T) {
	var addrs = UniqCSVAddrs(2181, "127.0.0.1", "", "127.0.0.2:111", "127.0.0.3:2181,127.0.0.4:2181", "中:文")
	if got, want := strings.Join(addrs, ","), "127.0.0.1:2181,127.0.0.2:111,127.0.0.3:2181,127.0.0.4:2181"; got != want {
		t.Errorf("UniqCSVAddrs = %s, want %s", got, want)
	}
}

// 端口为0或留空、IPv6带方括号时的补全，以及同一地址不同写法的去重
func TestUniqCSVAddrsEdges(t *testing.T) {
	for _, c := range []struct {
		port uint16
		in   []string
		want string
	}{
		{2181, []string{"[::1]:0", "127.0.0.1:0"}, "[::1]:2181,127.0.0.1:2181"},
		{2181, []string{"[::1]", "127.0.0.1:", "[::1]:"}, "[::1]:2181,127.0.0.1:2181"},
		{0, []string{"[::1]", "127.0.0.1:", ":0", "a:80"}, "::1,127.0.0.1,a:80"},
		{0, []string{"[2001:DB8::1]:0", "2001:DB8::1", "[::FFFF:1.2.3.4]:", "::ffff:1.2.3.4"}, "2001:db8::1,1.2.3.4"},
		{80, []string{"[2001:DB8::1]:080", "2001:db8::1", "1.2.3.4:0080", "1.2.3.4:80", "a:080", "a:80"}, "[2001:db8::1]:80,1.2.3.4:80,a:80"},
	} {
		if got := strings.Join(UniqCSVAddrs(c.port, c.in...), ","); got != c.want {
			t.Errorf("UniqCSVAddrs(%d, %q) = %s, want %s", c.port, c.in, got, c.want)
		}
	}
}

// 无主机的地址须保留端口，以免整条地址变为空串
func TestTrimPort(t *testing.T) {
	var got = TrimPort(80, ":80", "1.2.3.4:80", "[::1]:80", "a:8080", "a")
	if want := []string{":80", "1.2.3.4", "::1", "a:8080", "a"}; !slices.Equal(got, want) {
		t.Errorf("TrimPort = %q, want %q", got, want)
	}
	if got := PrettyCSVAddr(80, "1.2.3.4,:80"); got != "1.2.3.4,:80" {
		t.Errorf("PrettyCSVAddr = %q, want %q", got, "1.2.3.4,:80")
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
