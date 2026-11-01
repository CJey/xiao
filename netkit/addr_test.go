package netkit

import (
	"fmt"
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

func TestUniqCSVAddrsZeroPort(t *testing.T) {
	var got = strings.Join(UniqCSVAddrs(2181, "[::1]:0", "127.0.0.1:0"), ",")
	if want := "[::1]:2181,127.0.0.1:2181"; got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}
